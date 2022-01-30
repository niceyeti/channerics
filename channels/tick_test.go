// Copyright 2022 Jesse Waite

package channerics

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTick(t *testing.T) {
	maxWaitForEffect := time.Duration(50) * time.Millisecond

	Convey("Tick tests", t, func() {
		Convey("When done is already closed", func() {
			done := make(chan struct{})
			duration := time.Millisecond * 10
			close(done)
			ticker := NewTicker(done, duration)

			ok := true
			select {
			case _, ok = <-ticker:
			case <-time.After(maxWaitForEffect):
			}
			So(ok, ShouldBeFalse)
		})

		Convey("When tick has been sent but done is closed before it is read", func() {
			done := make(chan struct{})
			duration := time.Millisecond * 1
			ticker := NewTicker(done, duration)

			/// There is no way to deterministically synchronize with the sending of the
			// tick; this sleep ensures it almost certainly has been sent.
			<-time.After(duration * 20)
			close(done)

			// Drain the ticker; this must be done because there could be one pending
			// tick sent, non-deterministically chosen over the closure of 'done'.

			ok := true
			select {
			case _, ok = <-ticker:
			case <-time.After(maxWaitForEffect):
			}
			So(ok, ShouldBeFalse)
		})

		Convey("When tick is read -- happy path", func() {
			done := make(chan struct{})
			duration := time.Millisecond * 1
			ticker := NewTicker(done, duration)

			// Read a few ticks
			for i := 0; i < 4; i++ {
				select {
				case _, ok := <-ticker:
					So(ok, ShouldBeTrue)
				case <-time.After(duration * 5):
					t.FailNow()
				}
			}

			// ...then close.
			close(done)
			// Drain any pending tick. This must be done since since the inner select
			// to send the tick may be pending, and the select will non-deterministically
			// decide whether to send the tick or break on the closure of done.
			_ = <-ticker

			ok := true
			select {
			case _, ok = <-ticker:
			case <-time.After(maxWaitForEffect):
			}
			So(ok, ShouldBeFalse)
		})
	})
}
