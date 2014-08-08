package lib

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const cfg = `
{
    "root": "/proj",
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
		store := NewStore(NewConfig([]byte(cfg)))
		store.Put("/proj/app/controllers/foo.js", "foo")
		store.Put("/proj/app/models/bar.js", "bar")
		store.Put("/proj/app/routes/baz.js", "baz")
		store.Put("/proj/app/styles/app.css", "css")
		store.Put("/proj/app/styles/foo.css", "foocss")
		store.Put("/proj/app/templates/app.hbs", "hello")
		store.Put("/proj/vendor/qux/main.js", "qux")
		store.Put("/proj/vendor/bax/index.js", "bax")
		store.Put("/proj/vendor/qux/qux.css", "qstyle")

		Convey("A file's content can be retrieved", func() {
			So(store.Get("/proj/app/controllers/foo.js"), ShouldEqual, "foo")
		})

		Convey("A file's content can be updated", func() {
			store.Put("/proj/app/controllers/foo.js", "zzz")
			So(store.Get("/proj/app/controllers/foo.js"), ShouldEqual, "zzz")
		})

		Convey("A file's content can be deleted", func() {
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
				So(store.Get("app.js"), ShouldEqual, "app\nfoo\nbar\nbaz")
			})

			Convey("GetAll will include content for merge files", func() {
				f := store.GetAll()[10]
				So(f.Name, ShouldEqual, "app.js")
				So(f.Content, ShouldEqual, "foo\nbar\nbaz")
			})
		})

		Convey("app.css has the correct content", func() {
			So(store.Get("app.css"), ShouldEqual, "css\nfoocss")
		})

		Convey("vendor.js file has the correct content", func() {
			So(store.Get("vendor.js"), ShouldEqual, "qux\nbax")
		})

		Convey("It can retrieve all files ordered by name", func() {
			files := store.GetAll()

			So(files[0].Name, ShouldEqual, "/proj/app/controllers/foo.js")
			So(files[4].Name, ShouldEqual, "/proj/app/styles/foo.css")
		})
	})
}

func TestFileAccessors(t *testing.T) {
	Convey("Given files exist in the store", t, func() {
		store := NewStore(NewConfig([]byte(cfg)))
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
			f := store.GetFile("/proj/app/controllers/foo.js")
			So(f.Name, ShouldEqual, "/proj/app/controllers/foo.js")
			So(f.Content, ShouldEqual, "foo")
		})

		Convey("A defined file has the correct content", func() {
			f := store.GetFile("app.js")
			So(f.Name, ShouldEqual, "app.js")
			So(f.Type, ShouldEqual, "merge")
			So(f.Content, ShouldEqual, "foo\nbar\nbaz")
		})

	})
}

func TestUpdatedChannel(t *testing.T) {
	Convey("Given a file was added", t, func() {
		store := NewStore(NewConfig([]byte(cfg)))
		store.Put("/proj/app/controllers/foo.js", "foo")
		res := <-store.DidUpdate

		Convey("An update was triggered", func() {
			So(res.Name, ShouldEqual, "/proj/app/controllers/foo.js")
		})

		Convey("Deleting the file triggers an update", func() {
			store.Delete("/proj/app/controllers/foo.js")

			res = <-store.DidUpdate
			So(res.Name, ShouldEqual, "/proj/app/controllers/foo.js")
		})
	})
}

func TestStoreListen(t *testing.T) {
	Convey("Given a store is listening for files", t, func() {
		s := NewStore(NewConfig([]byte(cfg)))
		s.Listen()

		Convey("It handles file create", func() {
			s.Input <- NewFileWithContent("/proj/app/controllers/foo.js", "foo", CREATE)
			f := <-store.DidUpdate

			So(f.Name, ShouldEqual, "/proj/app/controllers/foo.js")
			So(f.Op, ShouldEqual, CREATE)
		})

		Convey("It handles file write", func() {
			s.Input <- NewFileWithContent("/proj/app/controllers/1.js", "1", WRITE)
			f := <-store.DidUpdate

			So(f.Name, ShouldEqual, "/proj/app/controllers/1.js")
			So(f.Op, ShouldEqual, WRITE)
		})

		Convey("It handles file removal", func() {
			s.Input <- NewFileWithContent("/proj/app/controllers/2.js", "", REMOVE)
			f := <-store.DidUpdate

			So(store.GetFile("/proj/app/controllers/2.js"), ShouldBeNil)
			So(f.Name, ShouldEqual, "/proj/app/controllers/2.js")
			So(f.Op, ShouldEqual, REMOVE)
		})

		Convey("It handles file rename", func() {
			s.Input <- NewFileWithContent("/proj/app/controllers/3.js", "", RENAME)
			f := <-store.DidUpdate

			So(store.GetFile("/proj/app/controllers/3.js"), ShouldBeNil)
			So(f.Name, ShouldEqual, "/proj/app/controllers/3.js")
			So(f.Op, ShouldEqual, RENAME)
		})

		Convey("It ignores other file modes", func() {
			s.Input <- NewFileWithContent("/proj/app/controllers/4.js", "", CHMOD)

			select {
			case f := <-store.DidUpdate:
				So("Fail - Store should not update", ShouldEqual, f)
			default:
				So("Pass - Store did not update", ShouldNotBeBlank)
			}
		})

		Convey("It ignores nil files", func() {
			var f *File
			s.Input <- f
			time.Sleep(time.Millisecond * 30)

			select {
			case f := <-store.DidUpdate:
				So("Fail - Store should not update", ShouldEqual, f)
			default:
				So("Pass - Store did not update", ShouldNotBeBlank)
			}
		})

		Convey("It ignores LOG files", func() {
			s.Input <- &File{Name: "foo", Op: LOG}
			time.Sleep(time.Millisecond * 30)

			select {
			case f := <-store.DidUpdate:
				So("Fail - Store should not update", ShouldEqual, f)
			default:
				So("Pass - Store did not update", ShouldNotBeBlank)
			}
		})
	})
}
