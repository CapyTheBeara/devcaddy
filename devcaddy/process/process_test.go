package process

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProcess(t *testing.T) {
	Convey("Node process", t, func() {
		p := NewProcess("node", "", "")

		Convey("Can process multiple inputs", func() {
			p.In <- `console.log('foo');`
			res := <-p.Out
			So(res.Error, ShouldBeNil)
			So(res.Content, ShouldEqual, "foo")

			p.In <- `console.log("bar\nfoo\n");`
			res = <-p.Out
			So(res.Error, ShouldBeNil)
			So(res.Content, ShouldEqual, "bar\nfoo")
		})

		Convey("Can process errors", func() {
			p.In <- "asfd"
			res := <-p.Out
			So(res.Content, ShouldEqual, "")
			So(res.Error.Error(), ShouldContainSubstring, "asfd is not defined")

			p.In <- "zzz"
			res = <-p.Out
			So(res.Content, ShouldEqual, "")
			So(res.Error.Error(), ShouldContainSubstring, "zzz is not defined")
		})

		Convey("Can accept a module for input processing", func() {
			testNodeModule := `exports.plugin = function(file, settings) {
                return {
                    name: file.name + settings.connector,
                    content: file.content + settings.connector
                };
            };`

			settings := `{ "connector": "!" }`

			p := NewProcess("node", testNodeModule, settings)
			p.In <- Marshal("foo.js", `bar . \
baz`)
			res := <-p.Out

			So(res.Error, ShouldBeNil)

			var jsn map[string]string
			err := json.Unmarshal([]byte(res.Content), &jsn)
			if err != nil {
				t.Fatal(err)
			}

			So(jsn["name"], ShouldEqual, "foo.js!")
			So(jsn["content"], ShouldEqual, `bar . \
baz!`)
		})
	})
}
