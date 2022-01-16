// Copyright 2022 Jesse Waite

package channerics

import (
	"testing"

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
