// Copyright 2022 Jesse Waite

package channerics

import (
	"context"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSemaphore(t *testing.T) {
	Convey("Semaphore tests", t, func() {
		Convey("When NewSemaphore is called -- for coverage", func() {
			sem := NewSemaphore(3)
			defer sem.Close()
			So(len(sem.ch), ShouldEqual, 0)
			So(cap(sem.ch), ShouldEqual, 3)
		})

		Convey("When Take is called", func() {
			sem := NewSemaphore(2)
			defer sem.Close()
			ctx := context.Background()
			taken := sem.Take(ctx)
			So(taken, ShouldBeTrue)
			taken = sem.Take(ctx)
			So(taken, ShouldBeTrue)

			// Count is zero, now we should block.
			blocked := make(chan struct{})
			isBlocked := false
			taken = false

			// Wait until the Take() is scheduled
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				wg.Done()
				taken = sem.Take(ctx)
				close(blocked)
			}()
			wg.Wait()

			select {
			case <-blocked:
			default:
				isBlocked = true
			}

			So(isBlocked, ShouldBeTrue)
			So(taken, ShouldBeFalse)

			// Release the semaphore; since there are no other users of it, Take() should be unblocked.
			sem.Release()
			// Wait or timeout for Take() to become unblocked
			select {
			case _, isBlocked = <-blocked:
			case <-time.After(time.Duration(250) * time.Millisecond):
			}

			So(isBlocked, ShouldBeFalse)
			So(taken, ShouldBeTrue)
		})

		Convey("When cancelled before Take", func() {
			sem := NewSemaphore(0)
			defer sem.Close()
			ctx, cancelFn := context.WithCancel(context.Background())
			// Cancel before Take()
			cancelFn()

			taken := false
			runner := func() {
				taken = sem.Take(ctx)
			}
			didComplete := RunOrTimeout(runner, 250)

			So(didComplete, ShouldBeFalse)
			So(taken, ShouldBeFalse)
		})

		Convey("When cancelled after Take", func() {
			sem := NewSemaphore(0)
			defer sem.Close()

			ctx, cancelFn := context.WithCancel(context.Background())
			takeResult := true
			takeCompleted := make(chan struct{})
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				wg.Done()
				takeResult = sem.Take(ctx)
				close(takeCompleted)
			}()
			// Wait until routine is running...
			wg.Wait()

			So(takeResult, ShouldBeTrue)

			// Cancel and wait until Take returns or timeout occurs.
			cancelFn()
			timedOut := false
			select {
			case <-takeCompleted:
			case <-time.After(time.Duration(250) * time.Millisecond):
				timedOut = true
			}

			So(timedOut, ShouldBeFalse)
			So(takeResult, ShouldBeFalse)
		})

		Convey("When UnsafeCount is called", func() {
			// Note: testing UnsafeCount values in this test is valid only because this is the only go routine accessing the semaphore.
			sem := NewSemaphore(10)
			defer sem.Close()

			So(sem.UnsafeCount(), ShouldEqual, 10)
			taken := sem.Take(context.Background())
			So(taken, ShouldBeTrue)
			So(sem.UnsafeCount(), ShouldEqual, 9)
		})

		Convey("When Release is called", func() {
			// Note: usage of UnsafeCount is valid in this test only because this is the only goroutine accessing the semaphore,
			// a degenerate usage of a semaphore. UnsafeCount should otherwise NOT be relied on in this manner.
			sem := NewSemaphore(10)
			defer sem.Close()

			So(sem.UnsafeCount(), ShouldEqual, 10)
			taken := sem.Take(context.Background())
			So(taken, ShouldBeTrue)
			So(sem.UnsafeCount(), ShouldEqual, 9)
			sem.Release()
			So(sem.UnsafeCount(), ShouldEqual, 10)
			// Ensure releasing again after reaching capacity has no effect.
			sem.Release()
			So(sem.UnsafeCount(), ShouldEqual, 10)
		})

		Convey("When MaybeTake is called", func() {
			// Note: usage of UnsafeCount is valid in this test only because this is the only goroutine accessing the semaphore,
			// a degenerate usage of a semaphore. UnsafeCount should otherwise NOT be relied on in this manner.
			sem := NewSemaphore(10)
			defer sem.Close()

			for i := 0; i < 10; i++ {
				taken := sem.MaybeTake()
				So(taken, ShouldBeTrue)
			}

			So(sem.UnsafeCount(), ShouldEqual, 0)
			So(sem.MaybeTake(), ShouldBeFalse)
			So(sem.UnsafeCount(), ShouldEqual, 0)

		})

	})
}

// RunOrTimeout ensures the passed function runs to completion before the passed timeout.
func RunOrTimeout(fn func(), timeoutMs int) (timedOut bool) {
	ch := make(chan struct{})
	go func() {
		fn()
		close(ch)
	}()

	select {
	case <-ch:
	case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
		timedOut = true
	}
	return
}

// WaitUntilSheduled waits until fn is scheduled and called; note this
// does not mean fn has completed. Returns true if fn was scheduled,
// false if cancelled via context.
func WaitUntilCalled(ctx context.Context, fn func()) (isLive bool) {
	waitChan := make(chan struct{})
	go func() {
		close(waitChan)
		fn()
	}()

	select {
	case <-waitChan:
		isLive = true
	case <-ctx.Done():
	}
	return
}
