package lib

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const cfg = `
{
    "root": "",
    "files": [
        {
            "name": "app.js",
            "dir": "app",
            "ext": "js",
            "type": "merge"
        },
        {
            "name": "app.css",
            "dir": "app",
            "ext": "css",
            "type": "merge"
        },
        {
            "name": "vendor.js",
            "dir": "vendor",
            "files": ["qux/main.js", "bax/index.js"],
            "type": "merge"
        }
    ]
}
`

func TestFileAccessing(t *testing.T) {
	Convey("Given files exist in the store", t, func() {
		store := NewStore("/proj", NewConfig([]byte(cfg)))
		store.Put("/proj/app/controllers/foo.js", "foo")
		store.Put("/proj/app/models/bar.js", "bar")
		store.Put("/proj/app/routes/baz.js", "baz")
		store.Put("/proj/app/styles/app.css", "css")
		store.Put("/proj/app/styles/foo.css", "foocss")
		store.Put("/proj/app/templates/app.hbs", "hello")
		store.Put("/proj/vendor/qux/main.js", "qux")
		store.Put("/proj/vendor/bax/index.js", "bax")
		store.Put("/proj/vendor/qux/qux.css", "qstyle")

		Convey("A file can be retrieved", func() {
			So(store.Get("/proj/app/controllers/foo.js"), ShouldEqual, "foo")
		})

		Convey("A file can be deleted", func() {
			store.Delete("/proj/app/routes/baz.js")
			So(store.Get("/proj/app/routes/baz.js"), ShouldEqual, "")
		})

		Convey("Given app.js file exists", func() {
			Convey("It has the correct content", func() {
				So(store.Get("app.js"), ShouldEqual, "foo\nbar\nbaz")
			})

			Convey("It has the correct content after deleting a file", func() {
				store.Delete("/proj/app/controllers/foo.js")
				So(store.Get("app.js"), ShouldEqual, "bar\nbaz")
			})

			Convey("It has the correct content after adding a file", func() {
				store.Put("/proj/app/controllers/app.js", "app")
				So(store.Get("app.js"), ShouldEqual, "foo\nbar\nbaz\napp")
			})
		})

		Convey("app.css has the correct content", func() {
			So(store.Get("app.css"), ShouldEqual, "css\nfoocss")
		})

		Convey("vendor.js file has the correct content", func() {
			So(store.Get("vendor.js"), ShouldEqual, "qux\nbax")
		})
	})
}

func TestUpdatedChannel(t *testing.T) {
	Convey("Given a file was added", t, func() {
		store := NewStore("/proj", NewConfig([]byte(cfg)))
		var res string

		store.Put("/proj/app/controllers/foo.js", "foo")

		select {
		case res = <-store.DidUpdate:
		}

		Convey("An update was triggered", func() {
			So(res, ShouldEqual, "/proj/app/controllers/foo.js")
		})

		Convey("Deleting the file triggers an update", func() {
			store.Delete("/proj/app/controllers/foo.js")

			select {
			case res = <-store.DidUpdate:
				So(res, ShouldEqual, "/proj/app/controllers/foo.js")
			}

		})
	})
}
