package lib

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	// "time"

	. "github.com/smartystreets/goconvey/convey"
)

func makeTmpDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "filestore")
	if err != nil {
		t.Fatalf("failed to create test directory: %s", err)
	}
	return dir
}

func makeTmpFile(t *testing.T, dir, name, content string) *os.File {
	path := filepath.Join(dir, name)

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create test file: %s", err)
	}

	f.WriteString(content)
	f.Sync()
	f.Close()
	return f
}

const cfg1 = `
{
    "fileDefs": [
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
		store := NewStore("/proj", strings.NewReader(cfg1))
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
		store := NewStore("/proj", strings.NewReader(cfg1))
		var res bool

		store.Put("/proj/app/controllers/foo.js", "foo")

		select {
		case res = <-store.DidUpdate:
		}

		Convey("An update was triggered", func() {
			So(res, ShouldBeTrue)
		})

		Convey("Deleting the file triggers an update", func() {
			store.Delete("/proj/app/controllers/foo.js")

			select {
			case res = <-store.DidUpdate:
				So(res, ShouldBeTrue)
			}

		})
	})
}

// func TestFileLoading(t *testing.T) {
// 	Convey("Given actual files exist", t, func() {
// 		testDir := makeTmpDir(t)
// 		f1 := makeTmpFile(t, testDir, "foo.js", "foo")
// 		f2 := makeTmpFile(t, testDir, "bar.css", "bar")

// 		Convey("Files can be directly stored", func() {
// 			config := Config{
// 				Dirs: []string{testDir},
// 			}
// 			store := NewStore([]*Config{&config})

// 			So(store.Get(f1.Name()), ShouldEqual, "foo")
// 			So(store.Get(f2.Name()), ShouldEqual, "bar")
// 		})

// Convey("Configuratin can filter files by extension", func() {
// 	config := Config{
// 		Ext:  "js",
// 		Dirs: []string{testDir},
// 	}
// 	store := NewStore([]*Config{&config})

// 	So(store.Get(f1.Name()), ShouldEqual, "foo")
// 	So(store.Get(f2.Name()), ShouldEqual, "")

// })

// Convey("Configuration can specify a processing function", func() {
// 	newPath := func(p string) string {
// 		return strings.Replace(p, ".js", ".JS", 1)
// 	}

// 	p := func(path, content string) (string, string) {
// 		return newPath(path), content + "!"
// 	}

// 	config := Config{
// 		Ext:       "js",
// 		Processor: p,
// 		Dirs:      []string{testDir},
// 	}
// 	store := NewStore([]*Config{&config})

// 	So(store.Get(f1.Name()), ShouldEqual, "")
// 	So(store.Get(newPath(f1.Name())), ShouldEqual, "foo!")
// 	So(store.Get(f2.Name()), ShouldEqual, "")
// })

// 		Reset(func() {
// 			os.RemoveAll(testDir)
// 		})
// 	})

// }
