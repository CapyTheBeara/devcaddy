package lib

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const cfg2 = `
{
    "root": "../mockapp",
    "plugins": [
        {
            "name": "transpile-js",
            "command": "echo",
            "args": "-n"
        },
        {
            "name": "silent",
            "command": "echo",
            "args": "silent",
            "noOutput": true
        },
        {
            "name": "lint",
            "command": "echo",
            "args": "lint",
            "logOnly": true
        },
        {
            "name": "template",
            "command": "echo",
            "args": "template",
            "pipeTo": "transpile-js"
        }
    ],
    "watch": [
        {
            "dir": "app/templates",
            "ext": "hbs",
            "plugins": ["template"]
        }
    ],
    "files": [
        {
            "name": "app.js",
            "dir": "app",
            "ext": "js",
            "plugins": ["transpile-js", "silent", "lint"]
        },
        {
            "name": "vendor.js",
            "dir": "vendor",
            "files": ["bar/index.js", "baz/main.js"]
        },
        {
            "name": "index.html",
            "dir": "app",
            "files": ["index.html"]
        }
    ]
}
`

func TestPluginCreation(t *testing.T) {
	Convey("Plugins from the config", t, func() {
		c := NewConfig([]byte(cfg2))
		p := c.Plugins[0]

		Convey("Names should be correct", func() {
			So(p.Name, ShouldEqual, "transpile-js")
		})

		Convey("Tranformers should be correct", func() {
			p.InC <- &File{Name: "foo.js", Content: "hello"}

			select {
			case res := <-p.OutC:
				So(res.Name, ShouldEqual, "foo.js")
				So(res.Content, ShouldEqual, "foo.js hello")
			}
		})
	})
}

func TestMakeStore(t *testing.T) {
	Convey("PopulateStore creates a store and adds initial files", t, func() {
		done := make(chan bool)
		c := NewConfig([]byte(cfg2))
		store := c.PopulateStore(done)

		<-done
		Convey("Defined files should be correct", func() {
			So(store.Get("app.js"), ShouldEqual, "../mockapp/app/foo.js var foo;\n\n../mockapp/app/main.js var main;\n")
			So(store.Get("vendor.js"), ShouldEqual, "var bar;\n\nvar baz;\n")
			So(store.Get("index.html"), ShouldEqual, "<html></html>\n")
		})

		Convey("Store should not contain unspecified files", func() {
			So(store.Get("../mockapp/vendor/bar/nope.js"), ShouldEqual, "")
			So(store.Get("../mockapp/app/nope.foo"), ShouldEqual, "")
		})

		Convey("Plugin with PipeTo set, has the correct content", func() {
			So(store.Get("../mockapp/app/templates/index.hbs"), ShouldEqual, "../mockapp/app/templates/index.hbs template ../mockapp/app/templates/index.hbs {{index}}\n\n")
		})

		Convey("After store is populated, watcher output goes to store", func() {
			updateTestFile(t, "../mockapp/app/foo.js", "")
			name := <-store.DidUpdate
			So(name, ShouldEqual, "../mockapp/app/foo.js")
		})

		Convey("Deleted files get removed from the store", func() {
			name := "../mockapp/app/arggg.js"
			makeTestFile(t, "../mockapp", "app/arggg.js", "arg!", 0)
			n := <-store.DidUpdate
			So(n, ShouldEqual, name)

			removeTestFile(t, name)
			<-store.DidUpdate
			_, ok := store.Files[name]
			So(ok, ShouldBeFalse)
		})
	})
}

func TestDefaultRoot(t *testing.T) {
	Convey("If no root specified in config.json, it sets it to the relative cwd", t, func() {
		c := NewConfig([]byte("{}"))
		So(c.Root, ShouldEqual, "../lib")
	})
}
