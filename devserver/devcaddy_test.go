package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/monocle/devcaddy/devcaddy/lib"
)

const cfg = `
{
    "plugins": [
        {
            "name": "transpile-js",
            "command": "echo",
            "args": "-n {{fileContent}}1"
        },
        {
            "name": "template",
            "command": "echo",
            "args": "-n {{fileContent}}2",
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
