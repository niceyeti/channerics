// Copyright 2022 Jesse Waite

package channerics

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOrDone(t *testing.T) {
	maxWaitForEffect := time.Duration(250) * time.Millisecond

	type foo struct {
		SomeString string
		SomeFloat  float64
	}

	Convey("OrDone Tests", t, func() {
		Convey("When one or more values are sent and read before done", func() {
			done := make(chan struct{})
			vals := make(chan *foo)
			orDone := OrDone(done, vals)

			go func() {
				vals <- &foo{SomeString: "bar", SomeFloat: 1234}
				vals <- &foo{SomeString: "baz", SomeFloat: 4321}
			}()

			var v1, v2 *foo
			var ok1, ok2 bool
			select {
			case v1, ok1 = <-orDone:
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}
			So(ok1, ShouldBeTrue)
			So(v1.SomeString, ShouldEqual, "bar")
			So(v1.SomeFloat, ShouldEqual, 1234)

			select {
			case v2, ok2 = <-orDone:
			case <-time.After(maxWaitForEffect):
			}
			So(ok2, ShouldBeTrue)
			So(v2.SomeString, ShouldEqual, "baz")
			So(v2.SomeFloat, ShouldEqual, 4321)
		})

		Convey("When done is already closed", func() {
			done := make(chan struct{})
			vals := make(chan *foo)
			close(done)
			orDone := OrDone(done, vals)

			closedViaDone := false
			select {
			case _, ok := <-orDone:
				closedViaDone = !ok
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}
			So(closedViaDone, ShouldBeTrue)
		})

		Convey("When vals are sent but done is closed before they are read", func() {
			done := make(chan struct{})
			vals := make(chan *foo)
			orDone := OrDone(done, vals)

			select {
			case vals <- &foo{SomeString: "baz", SomeFloat: 4321}:
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}
			// Wait a brief period for send to propagate to select stmt within OrDone.
			close(done)

			closedViaDone := false
			select {
			case _, ok := <-orDone:
				closedViaDone = !ok
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}
			So(closedViaDone, ShouldBeTrue)
		})

		Convey("When vals channel is closed before OrDone", func() {
			done := make(chan struct{})
			vals := make(chan *foo)
			close(vals)
			orDone := OrDone(done, vals)

			closedViaDone := false
			select {
			case _, ok := <-orDone:
				closedViaDone = !ok
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}
			So(closedViaDone, ShouldBeTrue)
		})

		Convey("When vals channel is closed after send but before read, output must be drained", func() {
			done := make(chan struct{})
			vals := make(chan *foo)
			orDone := OrDone(done, vals)

			sendCompleted := false
			select {
			case vals <- &foo{SomeString: "baz", SomeFloat: 4321}:
				sendCompleted = true
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}
			So(sendCompleted, ShouldBeTrue)

			close(vals)
			// Input is now closed, but output must be drained before detecting input closure.
			readCompleted := false
			select {
			case _, ok := <-orDone:
				readCompleted = ok
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}
			So(readCompleted, ShouldBeTrue)

			// Finally, orDone should be closed
			closedViaOrDone := false
			select {
			case _, ok := <-orDone:
				closedViaOrDone = !ok
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}
			So(closedViaOrDone, ShouldBeTrue)
		})
	})
}

func TestAny(t *testing.T) {
	maxWaitForEffect := time.Duration(100) * time.Millisecond

	Convey("Any tests", t, func() {
		Convey("When nil channel is passed", func() {
			var ch chan struct{}
			out := Any(ch)

			isBlocked := false
			select {
			case <-out:
			default:
				isBlocked = true
			}
			So(isBlocked, ShouldBeTrue)
		})

		Convey("When no channel is passed", func() {
			chans := []<-chan struct{}{}
			out := Any(chans...)

			isBlocked := false
			select {
			case <-out:
			default:
				isBlocked = true
			}
			So(isBlocked, ShouldBeTrue)
		})

		Convey("When two channels are passed", func() {
			ch1 := make(chan struct{})
			ch2 := make(chan struct{})
			out := Any(ch1, ch2)

			isBlocked := false
			select {
			case <-out:
			default:
				isBlocked = true
			}
			So(isBlocked, ShouldBeTrue)

			close(ch1)
			select {
			case <-out:
				isBlocked = false
			case <-time.After(time.Duration(250) * time.Millisecond):
				isBlocked = true
			}
			So(isBlocked, ShouldBeFalse)
		})

		Convey("When eitherDone is called -- for coverage", func() {
			Convey("When first channel is closed first", func() {
				ch1 := make(chan struct{})
				ch2 := make(chan struct{})
				done := eitherDone(ch1, ch2)

				close(ch1)
				exitedViaDone := false
				select {
				case <-done:
					exitedViaDone = true
				case <-time.After(maxWaitForEffect):
				}
				So(exitedViaDone, ShouldBeTrue)
			})

			Convey("When second channel is closed first", func() {
				ch1 := make(chan struct{})
				ch2 := make(chan struct{})
				done := eitherDone(ch1, ch2)

				close(ch2)
				exitedViaDone := false
				select {
				case <-done:
					exitedViaDone = true
				case <-time.After(maxWaitForEffect):
				}
				So(exitedViaDone, ShouldBeTrue)
			})

		})

		Convey("When several channels are passed", func() {
			for i := 0; i < 4; i++ {
				chans := []chan struct{}{
					make(chan struct{}),
					make(chan struct{}),
					make(chan struct{}),
					make(chan struct{}),
				}

				inChans := make([]<-chan struct{}, len(chans))
				for i, ch := range chans {
					inChans[i] = ch
				}

				out := Any(inChans...)
				close(chans[i])
				time.Sleep(time.Duration(25) * time.Millisecond)

				closedViaDone := false
				closedViaTimeout := false
				select {
				case _, ok := <-out:
					closedViaDone = !ok
				case <-time.After(time.Duration(250) * time.Millisecond):
					closedViaTimeout = true
				}

				So(closedViaDone, ShouldBeTrue)
				So(closedViaTimeout, ShouldBeFalse)
			}
		})
	})
}

func TestAll(t *testing.T) {
	Convey("All tests", t, func() {
		Convey("When chans is empty", func() {
			var chans []<-chan struct{}
			done := make(chan struct{})
			out := All(done, chans...)

			// All's worker is not immediately scheduled, so we must await closure briefly.
			closedAsExpected := false
			select {
			case <-out:
				closedAsExpected = true
			case <-time.After(time.Duration(50) * time.Millisecond):
			}

			So(closedAsExpected, ShouldBeTrue)
		})

		Convey("When done is closed before All none of the input chans are awaited", func() {
			chans := []chan struct{}{
				make(chan struct{}),
				make(chan struct{}),
			}
			inChans := make([]<-chan struct{}, len(chans))
			for i, ch := range chans {
				inChans[i] = ch
			}
			done := make(chan struct{})
			out := All(done, inChans...)

			close(done)
			closedAsExpected := false
			select {
			case _, ok := <-out:
				closedAsExpected = !ok
			// All's worker is not immediately scheduled, so we must await closure briefly.
			case <-time.After(time.Duration(50) * time.Millisecond):
			}

			So(closedAsExpected, ShouldBeTrue)
		})

		Convey("When chans is non-empty -- happy path", func() {
			chans := []chan struct{}{
				make(chan struct{}),
				make(chan struct{}),
			}
			inChans := make([]<-chan struct{}, len(chans))
			for i, ch := range chans {
				inChans[i] = ch
			}
			inDone := make(chan struct{})
			allDone := All(inDone, inChans...)

			// Closing one of the two chans allows out to continue blocking.
			close(chans[0])
			stillBlocking := false
			select {
			case <-allDone:
			case <-time.After(time.Duration(10) * time.Millisecond):
				stillBlocking = true
			}
			So(stillBlocking, ShouldBeTrue)

			// Closing the second chan should mean all chans are closed, and done no longer blocks.
			close(chans[1])
			closedViaDone := false
			select {
			case _, ok := <-allDone:
				closedViaDone = !ok
			case <-time.After(time.Duration(50) * time.Millisecond):
			}
			So(closedViaDone, ShouldBeTrue)
		})
	})
}
