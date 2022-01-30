// Copyright 2022 Jesse Waite

package channerics

import (
	"fmt"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConverters(t *testing.T) {
	Convey("AsType tests", t, func() {
		Convey("AsType is called when done already closed", func() {
			done := make(chan struct{})
			in := make(chan interface{})

			close(done)
			var out <-chan string = AsType[string](done, in)

			isOpen := true
			select {
			case _, isOpen = <-out:
			case <-time.After(time.Duration(250) * time.Millisecond):
				t.FailNow()
			}
			So(isOpen, ShouldBeFalse)
		})

		Convey("AsType is called when done closed between send and receive", func() {
			done := make(chan struct{})
			in := make(chan interface{})
			var out <-chan string = AsType[string](done, in)

			// Await send
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				wg.Done()
				in <- interface{}("string")
			}()
			wg.Wait()

			// Wait a trivial period for input to propagate to the output select stmt.
			time.Sleep(time.Duration(50) * time.Millisecond)
			close(done)

			closedViaDone := false
			select {
			case _, ok := <-out:
				closedViaDone = !ok
			case <-time.After(time.Duration(250) * time.Millisecond):
				t.FailNow()
			}

			So(closedViaDone, ShouldBeTrue)
		})

		Convey("AsType is called and values are sent and received -- happy path", func() {
			done := make(chan struct{})
			in := make(chan interface{})
			var out <-chan string = AsType[string](done, in)

			// Await send
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				wg.Done()
				in <- interface{}("success")
			}()
			wg.Wait()

			ok := false
			val := ""
			select {
			case val, ok = <-out:
			case <-time.After(time.Duration(250) * time.Millisecond):
				t.FailNow()
			}
			So(ok, ShouldBeTrue)
			So(val, ShouldEqual, "success")
		})

		Convey("AsType is called when input is already closed", func() {
			done := make(chan struct{})
			in := make(chan interface{})

			close(in)
			var out <-chan string = AsType[string](done, in)

			isOpen := true
			select {
			case _, isOpen = <-out:
			case <-time.After(time.Duration(250) * time.Millisecond):
				t.FailNow()
			}
			So(isOpen, ShouldBeFalse)
		})
	})

	Convey("Adapter tests", t, func() {
		Convey("Adapter is called when done already closed", func() {
			done := make(chan struct{})
			in := make(chan int)

			close(done)
			var out <-chan string = Adapter(done, in, func(i int) string { return fmt.Sprint(i) })

			isOpen := true
			select {
			case _, isOpen = <-out:
			case <-time.After(time.Duration(250) * time.Millisecond):
				t.FailNow()
			}
			So(isOpen, ShouldBeFalse)
		})

		Convey("Adapter is called when done closed between send and receive", func() {
			done := make(chan struct{})
			in := make(chan int)

			var out <-chan string = Adapter(done, in, func(i int) string { return fmt.Sprint(i) })

			// Await send
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				wg.Done()
				in <- 123
			}()
			wg.Wait()

			// Wait a trivial period for input to propagate to the output select stmt.
			time.Sleep(time.Duration(50) * time.Millisecond)
			close(done)

			outClosed := false
			select {
			case _, ok := <-out:
				outClosed = !ok
			case <-time.After(time.Duration(250) * time.Millisecond):
				t.FailNow()
			}

			So(outClosed, ShouldBeTrue)
		})

		Convey("Adatper is called and values are sent and received -- happy path", func() {
			done := make(chan struct{})
			in := make(chan int)

			var out <-chan string = Adapter(done, in, func(i int) string { return fmt.Sprint(i) })

			// Await send
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				wg.Done()
				in <- 123
			}()
			wg.Wait()

			ok := false
			val := ""
			select {
			case val, ok = <-out:
			case <-time.After(time.Duration(250) * time.Millisecond):
				t.FailNow()
			}
			So(ok, ShouldBeTrue)
			So(val, ShouldEqual, "123")
		})

		Convey("Adapter is called when input is already closed", func() {
			done := make(chan struct{})
			in := make(chan int)

			close(in)
			var out <-chan string = Adapter(done, in, func(i int) string { return fmt.Sprint(i) })

			isOpen := true
			select {
			case _, isOpen = <-out:
			case <-time.After(time.Duration(250) * time.Millisecond):
				t.FailNow()
			}
			So(isOpen, ShouldBeFalse)
		})
	})

}
