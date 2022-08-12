// Copyright Jesse Waite 2022

package channerics

import (
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBroadcast(t *testing.T) {
	Convey("Merge Tests", t, func() {
		Convey("When done is already closed before sending any values", func() {
			vals := make(chan int)
			done := make(chan struct{})
			close(done)
			outputs := Broadcast(done, vals, 3)

			So(len(outputs), ShouldEqual, 3)
			for i := 0; i < 3; i++ {
				_, ok := <-outputs[i]
				So(ok, ShouldBeFalse)
			}
		})

		Convey("When done is closed before sending any values", func() {
			vals := make(chan int)
			done := make(chan struct{})
			outputs := Broadcast(done, vals, 3)
			So(len(outputs), ShouldEqual, 3)
			close(done)

			for i := 0; i < 3; i++ {
				_, ok := <-outputs[i]
				So(ok, ShouldBeFalse)
			}
		})

		Convey("When one value is sent and then done is closed", func() {
			vals := make(chan int)
			done := make(chan struct{})
			outputs := Broadcast(done, vals, 3)
			So(len(outputs), ShouldEqual, 3)

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				wg.Done()
				vals <- 123
			}()
			wg.Wait()

			for _, output := range outputs {
				val, ok := <-output
				So(ok, ShouldBeTrue)
				So(val, ShouldEqual, 123)
			}

			wg.Add(1)
			go func() {
				wg.Done()
				vals <- 456
			}()
			wg.Wait()

			close(done)
			for i := 0; i < 3; i++ {
				_, ok := <-outputs[i]
				So(ok, ShouldBeFalse)
			}
		})
	})
}
