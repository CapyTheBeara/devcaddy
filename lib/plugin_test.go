package lib

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func createFile() *File {
	return &File{Name: "foo.js", Content: "hello"}
}

func TestPlugin(t *testing.T) {
	inputFile := createFile()

	Convey("Given a Plugin with a transformer function", t, func() {

		fn := func(f *File) *File {
			return &File{Name: f.Name + "1", Content: f.Content + "!"}
		}

		Convey("It can manipulate it's input", func() {
			p := NewPlugin(&PluginConfig{}, fn)
			p.InC <- inputFile
			res := <-p.OutC

			So(res.Name, ShouldEqual, "foo.js1")
			So(res.Content, ShouldEqual, "hello!")
		})

		Convey("It can be configured to not send it's transformer output", func() {
			p := NewPlugin(&PluginConfig{NoOutput: true}, fn)
			p.InC <- inputFile
			res := <-p.OutC

			So(res, ShouldBeNil)
		})
	})
}

func TestCommandPlugin(t *testing.T) {
	inputFile := createFile()

	Convey("Given a CommandPlugin", t, func() {
		Convey("An input file's name and content is sent as params", func() {
			p := NewCommandPlugin(&PluginConfig{
				Command: "echo",
				Args:    "-n",
			})
			p.InC <- inputFile
			res := <-p.OutC

			So(res.Name, ShouldEqual, "foo.js")
			So(res.Content, ShouldEqual, "foo.js hello")
		})

		Convey("Command error is added to the output", func() {
			p := NewCommandPlugin(&PluginConfig{
				Command: "node",
				Args:    "aasdfssdf.js",
			})
			p.InC <- inputFile
			res := <-p.OutC

			So(res.Name, ShouldEqual, "foo.js")
			So(res.Content, ShouldContainSubstring, "Cannot find module")
			So(res.Error, ShouldNotBeNil)
		})

		Convey("Command can change the output file's name", func() {
			p := NewCommandPlugin(&PluginConfig{
				Command: "echo",
				Args:    "-n __SERVER_FILE_PATH__=",
			})
			p.InC <- inputFile
			res := <-p.OutC

			So(res.Name, ShouldEqual, "foo.js hello")
			So(res.Content, ShouldEqual, "")
		})

		Convey("Args can specify whether file path or content is sent - ", func() {
			Convey("Only file path can be sent", func() {
				p := NewCommandPlugin(&PluginConfig{
					Command: "echo",
					Args:    "-n {{fileName}}",
				})
				p.InC <- inputFile
				res := <-p.OutC

				So(res.Name, ShouldEqual, "foo.js")
				So(res.Content, ShouldEqual, "foo.js")
			})

			Convey("Only file content can be sent", func() {
				p := NewCommandPlugin(&PluginConfig{
					Command: "echo",
					Args:    "-n {{fileContent}}",
				})
				p.InC <- inputFile
				res := <-p.OutC

				So(res.Name, ShouldEqual, "foo.js")
				So(res.Content, ShouldEqual, "hello")
			})

			Convey("Only additional args can be sent after {{}}", func() {
				p := NewCommandPlugin(&PluginConfig{
					Command: "echo",
					Args:    "-n {{fileContent}} bar",
				})
				p.InC <- inputFile
				res := <-p.OutC

				So(res.Name, ShouldEqual, "foo.js")
				So(res.Content, ShouldEqual, "hello bar")
			})
		})
	})
}
