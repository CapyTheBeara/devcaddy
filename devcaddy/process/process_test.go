package process

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const testDelim = "__DEVCADDY_END__"
const testNodeModule = `exports.plugin = function(input) {
   return input+'!';
}`
const testNodeModule2 = `exports.plugin = function(input) {
   return input+'?';
}`

func TestProcess(t *testing.T) {
	Convey("Node process", t, func() {
		p := NewProcess("node", testDelim, "")

		Convey("Can process multiple inputs", func() {
			p.In <- `console.log('foo');` + testDelim
			res := <-p.Out
			So(res.Content, ShouldEqual, "foo")
			So(res.Error, ShouldBeNil)

			p.In <- `console.log("bar\nfoo\n");` + testDelim
			res = <-p.Out
			So(res.Content, ShouldEqual, "bar\nfoo")
			So(res.Error, ShouldBeNil)
		})

		Convey("Can process errors", func() {
			p.In <- "asfd" + testDelim
			res := <-p.Out
			So(res.Content, ShouldEqual, "")
			So(res.Error.Error(), ShouldContainSubstring, "asfd is not defined")

			p.In <- "zzz" + testDelim
			res = <-p.Out
			So(res.Content, ShouldEqual, "")
			So(res.Error.Error(), ShouldContainSubstring, "zzz is not defined")
		})

		Convey("Can accept a module for input processing", func() {
			p := NewProcess("node", testDelim, testNodeModule)
			p.In <- "fooz\nbarz" + testDelim
			res := <-p.Out
			So(res.Content, ShouldEqual, "fooz\nbarz!")
			So(res.Error, ShouldBeNil)
		})
	})
}
