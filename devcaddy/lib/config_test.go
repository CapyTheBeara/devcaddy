package lib

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewConfig(t *testing.T) {
	Convey("If no root specified in config.json, it sets it to the relative cwd", t, func() {
		c := NewConfig([]byte("{}"))
		So(c.Root, ShouldEqual, "../lib")
	})

	Convey("It sets all the File object's Type to merge", t, func() {
		c := NewConfig([]byte(`{ "files": [{ "name": "foo" }] }`))

		So(c.Files[0].Type, ShouldEqual, "merge")
	})
}
