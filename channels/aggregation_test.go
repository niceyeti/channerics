// Copyright 2022 Jesse Waite

package channerics

import (
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMerge(t *testing.T) {
	maxWaitForEffect := time.Duration(250) * time.Millisecond

	Convey("Merge Tests", t, func() {
		Convey("When done is already closed before reading any value", func() {
			done := make(chan struct{})
			vals1 := make(chan string)
			vals2 := make(chan string)
			close(done)

			merged := Merge(done, vals1, vals2)

			chanClosed := false
			select {
			case _, ok := <-merged:
				chanClosed = !ok
			case <-time.After(maxWaitForEffect):
			}

			So(chanClosed, ShouldBeTrue)
		})

		Convey("When done is closed after send, before read", func() {
			done := make(chan struct{})
			vals1 := make(chan string)
			vals2 := make(chan string)
			merged := Merge(done, vals1, vals2)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				vals1 <- "abc"
				wg.Done()
			}()
			wg.Wait()

			// We must wait for the value to propagate to the output select stmt.
			time.Sleep(time.Duration(50) * time.Millisecond)
			close(done)
			// Wait again before reading for 'done' closure to propagate; otherwise
			// the read from output will be ambiguous, as the sending select stmt in
			// Merge will randomly decide between the send of done cases.
			time.Sleep(time.Duration(50) * time.Millisecond)

			chanClosed := false
			select {
			case _, ok := <-merged:
				chanClosed = !ok
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}

			So(chanClosed, ShouldBeTrue)
		})

		Convey("When inputs are already closed before reading any value", func() {
			done := make(chan struct{})
			vals1 := make(chan string)
			vals2 := make(chan string)
			close(vals1)
			close(vals2)

			merged := Merge(done, vals1, vals2)

			chanClosed := false
			select {
			case _, ok := <-merged:
				chanClosed = !ok
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}

			So(chanClosed, ShouldBeTrue)
		})

		Convey("When inputs are closed after send, before read", func() {
			done := make(chan struct{})
			vals1 := make(chan string)
			vals2 := make(chan string)
			merged := Merge(done, vals1, vals2)

			valsSent := make(chan struct{})
			go func() {
				vals1 <- "abc"
				vals2 <- "def"
				close(valsSent)
			}()
			// Await vals to be sent
			sent := false
			select {
			case <-valsSent:
				sent = true
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}
			So(sent, ShouldBeTrue)

			close(vals1)
			close(vals2)
			// We must wait for 'done' closure to propagate to the select stmt inside Merge.
			time.Sleep(time.Duration(50) * time.Millisecond)

			// Drain all output, so the inputs' closure can be detected.
			valsRead := make(chan struct{})
			go func() {
				<-merged
				<-merged
				close(valsRead)
			}()
			// Await vals to be sent
			rxed := false
			select {
			case <-valsRead:
				rxed = true
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}
			So(rxed, ShouldBeTrue)

			chanClosed := false
			select {
			case _, ok := <-merged:
				chanClosed = !ok
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}

			So(chanClosed, ShouldBeTrue)
		})

		Convey("When multiple values are sent -- happy path", func() {
			done := make(chan struct{})
			vals1 := make(chan string)
			vals2 := make(chan string)
			defer close(vals1)
			defer close(vals2)
			merged := Merge(done, vals1, vals2)

			// Send first value
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				wg.Done()
				vals1 <- "abc"
			}()
			wg.Wait()

			// Send second value
			wg.Add(1)
			go func() {
				wg.Done()
				vals2 <- "def"
			}()
			wg.Wait()

			ok := false
			val := ""
			select {
			case val, ok = <-merged:
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}
			So(ok, ShouldBeTrue)

			// Verify val; note there is no promise that output arrives in
			// the order sent, only that both are received.
			var nextVal string
			if val == "abc" {
				So(val, ShouldEqual, "abc")
				nextVal = "def"
			} else {
				So(val, ShouldEqual, "def")
				nextVal = "abc"
			}

			select {
			case val, ok = <-merged:
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}
			So(ok, ShouldBeTrue)
			So(val, ShouldEqual, nextVal)
		})
	})
}
