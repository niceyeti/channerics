// Copyright 2022 Jesse Waite

package channerics

import (
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerate(t *testing.T) {
	maxWaitForEffect := time.Duration(50) * time.Millisecond

	type foo struct {
		SomeString string
		SomeFloat  float64
	}

	Convey("Generate tests", t, func() {
		Convey("When done is already closed", func() {
			i := 0
			fn := func() (int, bool) {
				i++
				return i, true
			}
			done := make(chan struct{})
			close(done)
			gen := Generator(done, fn)

			closedViaDone := false
			select {
			case _, ok := <-gen:
				closedViaDone = !ok
			case <-time.After(maxWaitForEffect):
			}
			So(closedViaDone, ShouldBeTrue)
		})

		Convey("When gen is called once then returns false -- happy path", func() {
			i := 0
			fn := func() (int, bool) {
				i++
				if i == 1 {
					return 1, true
				}
				return i, false
			}
			done := make(chan struct{})
			gen := Generator(done, fn)

			ok := false
			val := 0
			select {
			case val, ok = <-gen:
			case <-time.After(maxWaitForEffect):
			}
			So(ok, ShouldBeTrue)
			So(val, ShouldEqual, 1)

			ok = false
			val = 0
			select {
			case val, ok = <-gen:
			case <-time.After(maxWaitForEffect):
			}
			So(ok, ShouldBeFalse)
			So(val, ShouldEqual, 0)
		})

		Convey("When gen is called and done is closed before it returns", func() {
			done := make(chan struct{})
			fn := func() (int, bool) {
				close(done)
				return 42, true
			}
			gen := Generator(done, fn)

			ok := true
			select {
			case _, ok = <-gen:
			case <-time.After(maxWaitForEffect):
			}
			// Note: this works because the closing of a channel occurs before any
			// receive that returns a zero value because the channel was closed.
			So(ok, ShouldBeFalse)
		})

		Convey("When gen is called and done is closed before reading output", func() {
			done := make(chan struct{})
			fn := func() (int, bool) {
				return 42, true
			}
			gen := Generator(done, fn)

			// Wait a brief period for output to propagate to sending-select statement in Generate.
			time.Sleep(25 * time.Millisecond)
			close(done)
			ok := true
			select {
			case _, ok = <-gen:
			case <-time.After(maxWaitForEffect):
			}
			// Note: this works because the closing of a channel occurs before any
			// receive that returns a zero value because the channel was closed.
			So(ok, ShouldBeFalse)
		})

	})
}

func TestRepeater(t *testing.T) {
	maxWaitForEffect := time.Duration(100) * time.Millisecond

	Convey("Test Repeater", t, func() {
		Convey("When done is already closed", func() {
			done := make(chan struct{})
			close(done)
			vals := []string{"abc", "def", "xyz"}
			out := Repeater(done, vals)

			ok := true
			select {
			case _, ok = <-out:
			case <-time.After(maxWaitForEffect):
			}
			So(ok, ShouldBeFalse)
		})

		Convey("When repeater runs then closes later", func() {
			done := make(chan struct{})
			vals := []string{"abc", "def"}
			out := Repeater(done, vals)

			for i := 0; i < 4; i++ {
				expectedVal := vals[i%len(vals)]
				val := ""
				ok := false
				select {
				case val, ok = <-out:
				case <-time.After(maxWaitForEffect):
				}
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, expectedVal)
			}

			// Wait for next value to propagate to the sending-select stmt in Repeater.
			<-time.After(maxWaitForEffect)
			close(done)

			ok := true
			select {
			case _, ok = <-out:
			case <-time.After(maxWaitForEffect):
			}
			So(ok, ShouldBeFalse)
		})

	})
}

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
