// Copyright 2022 Jesse Waite

package channerics

import (
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTee(t *testing.T) {

	type foo struct {
		SomeString string
		SomeFloat  float64
	}

	Convey("Tee tests", t, func() {
		Convey("When done is already closed", func() {
			f1 := &foo{
				SomeString: "foo 1",
				SomeFloat:  0.42,
			}
			in := make(chan *foo)
			defer close(in)
			done := make(chan struct{})
			close(done)

			out1, out2 := Tee(done, in)

			select {
			case in <- f1:
			case <-done:
			}

			v1, ok1 := <-out1
			So(ok1, ShouldBeFalse)
			So(v1, ShouldBeNil)

			v2, ok2 := <-out2
			So(ok2, ShouldBeFalse)
			So(v2, ShouldBeNil)
		})

		Convey("When input channel is already closed", func() {
			in := make(chan *foo)
			close(in)
			done := make(chan struct{})
			defer close(done)

			out1, out2 := Tee(done, in)

			v1, ok1 := <-out1
			So(ok1, ShouldBeFalse)
			So(v1, ShouldBeNil)

			v2, ok2 := <-out2
			So(ok2, ShouldBeFalse)
			So(v2, ShouldBeNil)
		})

		Convey("When done is closed after sending one item but not receiving -- for coverage", func() {
			f1 := &foo{
				SomeString: "foo 1",
				SomeFloat:  0.42,
			}
			in := make(chan *foo)
			done := make(chan struct{})

			out1, out2 := Tee(done, in)

			// Send one item
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				wg.Done()
				in <- f1
			}()
			wg.Wait()
			// Await propagation of the sent value a small period, before closing.
			time.Sleep(time.Millisecond)
			close(done)

			v1, ok1 := <-out1
			So(ok1, ShouldBeFalse)
			So(v1, ShouldBeNil)

			v2, ok2 := <-out2
			So(ok2, ShouldBeFalse)
			So(v2, ShouldBeNil)
		})

		Convey("When operating normally two sent values are received on both channels -- happy path", func() {
			f1 := &foo{
				SomeString: "foo 1",
				SomeFloat:  0.42,
			}
			f2 := &foo{
				SomeString: "foo 2",
				SomeFloat:  8675309,
			}
			in := make(chan *foo)
			defer close(in)
			done := make(chan struct{})
			defer close(done)

			out1, out2 := Tee(done, in)

			// Send two values
			go func() {
				in <- f1
				in <- f2
			}()

			v1, ok1 := <-out1
			So(ok1, ShouldBeTrue)
			So(v1.SomeString, ShouldEqual, "foo 1")

			v2, ok2 := <-out2
			So(ok2, ShouldBeTrue)
			So(v2.SomeString, ShouldEqual, "foo 1")

			v1, ok1 = <-out1
			So(ok1, ShouldBeTrue)
			So(v1.SomeString, ShouldEqual, "foo 2")

			v2, ok2 = <-out2
			So(ok2, ShouldBeTrue)
			So(v2.SomeString, ShouldEqual, "foo 2")
		})
	})
}
