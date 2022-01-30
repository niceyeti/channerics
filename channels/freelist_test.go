// Copyright 2022 Jesse Waite

package channerics

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFreeList(t *testing.T) {
	maxWaitForEffect := time.Duration(250) * time.Millisecond
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

		Convey("When we initialize a valid free list -- happy path", func() {
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

			var ok bool
			var result int
			// Get first item; should not block.
			select {
			default:
				result, ok = fl.Get()
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}
			So(result, ShouldEqual, 1)
			So(ok, ShouldBeTrue)

			ok = false
			// Get second item; should create a new item.
			select {
			default:
				result, ok = fl.Get()
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}

			So(result, ShouldEqual, 2)
			So(ok, ShouldBeTrue)

			// Now put em back
			ok = fl.Put(42)
			So(ok, ShouldBeTrue)
			ok = fl.Put(8675309)
			So(ok, ShouldBeFalse)
		})

		Convey("When Put and Get are called -- happy path", func() {
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

			var ok bool
			var result int
			// Get first item; should not block.
			select {
			default:
				result, ok = fl.Get()
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}
			So(result, ShouldEqual, 1)
			So(ok, ShouldBeTrue)

			// Put item back
			wasPut := fl.Put(result)
			So(wasPut, ShouldBeTrue)

			ok = false
			// Get second item; should not create a new item, since we put an item back.
			select {
			default:
				result, ok = fl.Get()
			case <-time.After(maxWaitForEffect):
				t.FailNow()
			}
			So(result, ShouldEqual, 1)
			So(ok, ShouldBeFalse)

			// Now put em back
			ok = fl.Put(42)
			So(ok, ShouldBeTrue)
			ok = fl.Put(8675309)
			So(ok, ShouldBeFalse)
		})

	})
}
