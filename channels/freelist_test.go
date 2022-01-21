// Copyright 2022 Jesse Waite

package channerics

import (
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFreeList(t *testing.T) {

	type foo struct {
		SomeString string
		SomeFloat  float64
	}

	Convey("FreeList tests", t, func() {
		Convey("When we try to init a FreeList of len 0", func() {
			fl, err := NewFreeList(
				0,
				func() int {
					return 2
				},
			)
			So(fl, ShouldBeNil)
			So(err, ShouldBeError, ErrInvalidSize)
		})

		Convey("When we initialize a valid free list", func() {
			var i int
			fl, err := NewFreeList(
				1,
				func() int {
					i++
					return i
				},
			)
			So(err, ShouldBeNil)
			So(fl, ShouldNotBeNil)

			var wg sync.WaitGroup
			var ok bool
			var result int
			wg.Add(1)
			go func() {
				// Get or timeout so test cannot hang if some change/bug makes us block.
				select {
				default:
					result, ok = fl.Get()
				case <-time.After(time.Duration(250) * time.Millisecond):
				}
				wg.Done()
			}()
			wg.Wait()

			So(result, ShouldEqual, 1)
			So(ok, ShouldBeTrue)

			wg.Add(1)
			ok = false
			go func() {
				// Get or timeout so test cannot hang if some change/bug makes us block.
				select {
				default:
					result, ok = fl.Get()
				case <-time.After(time.Duration(250) * time.Millisecond):
				}
				wg.Done()
			}()
			wg.Wait()

			So(result, ShouldEqual, 2)
			So(ok, ShouldBeTrue)

			// Now put em back
			ok = fl.Put(42)
			So(ok, ShouldBeTrue)
			ok = fl.Put(8675309)
			So(ok, ShouldBeFalse)

		})
	})
}
