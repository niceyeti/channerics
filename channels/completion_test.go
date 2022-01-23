// Copyright 2022 Jesse Waite

package channerics

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOrDone(t *testing.T) {

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

			v1, ok1 := <-orDone
			So(ok1, ShouldBeTrue)
			So(v1.SomeString, ShouldEqual, "bar")
			So(v1.SomeFloat, ShouldEqual, 1234)

			v2, ok2 := <-orDone
			So(ok2, ShouldBeTrue)
			So(v2.SomeString, ShouldEqual, "baz")
			So(v2.SomeFloat, ShouldEqual, 4321)
		})

		/*
		   Convey("When no values are sent before done", func(){

		   })

		   Convey("When one unread value is sent before done", func(){

		   })
		*/
	})
}

func TestAny(t *testing.T) {

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
			out := All(chans...)

			// All's worker is not immediately scheduled, so we must await closure briefly.
			closedAsExpected := false
			select {
			case <-out:
				closedAsExpected = true
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

			done := All(inChans...)

			// Closing one of the two chans allows out to continue blocking.
			close(chans[0])
			// Gives a momentary pause for closure to propagate across go routines.
			time.Sleep(time.Duration(10) * time.Millisecond)
			stillBlocking := false
			select {
			case <-done:
			default:
				stillBlocking = true
			}
			So(stillBlocking, ShouldBeTrue)

			// Closing the second chan should mean all chans are closed, and done no longer blocks.
			close(chans[1])
			closedViaDone := false
			select {
			case _, ok := <-done:
				closedViaDone = !ok
			case <-time.After(time.Duration(50) * time.Millisecond):
			}

			So(closedViaDone, ShouldBeTrue)
		})
	})
}
