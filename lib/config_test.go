package lib

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const cfg2 = `
{
    "plugins": [
        {
            "name": "transpile-js",
            "command": "echo",
            "args": "-n"
        },
        {
            "name": "lint",
            "command": "echo",
            "args": "lint",
            "noOutput": true
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
            "dir": "app",
            "ext": "js",
            "plugins": ["transpile-js"]
        },
        {
            "dir": "app/templates",
            "ext": "hbs",
            "plugins": ["template"]
        },
        {
            "dir": "vendor",
            "files": ["bar/index.js", "baz/main.js"]
        },
        {
            "dir": "app",
            "files": ["index.html"]
        }
    ],
    "files": [
        {
            "name": "app.js",
            "dir": "app",
            "ext": "js"
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

func TestProcessorCreation(t *testing.T) {
	Convey("Processors from the config", t, func() {
		c := NewConfig(cfg2, "../mockapp")
		p := c.Processors[0]

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
		c := NewConfig(cfg2, "../mockapp")
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
	})
}
