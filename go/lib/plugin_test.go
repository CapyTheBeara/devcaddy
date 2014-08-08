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

		Convey("If LogOnly is set, it will set the outpuf file mode to LOG", func() {
			p := NewPlugin(&PluginConfig{LogOnly: true}, fn)
			p.InC <- inputFile
			res := <-p.OutC

			So(res.Op, ShouldEqual, LOG)
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

		Convey("File shouldn't be processed if it was deleted", func() {
			p := NewCommandPlugin(&PluginConfig{
				Command: "echo",
				Args:    "-n {{fileContent}} bar",
			})
			inputFile.Op = REMOVE
			p.InC <- inputFile
			res := <-p.OutC

			So(res.Name, ShouldEqual, "foo.js")
			So(res.Content, ShouldEqual, "")
			So(res.Op, ShouldEqual, REMOVE)
		})
	})
}

func TestNewPlugins(t *testing.T) {
	Convey("creatPlugins correcly creates the Plugins", t, func() {
		pcs := []*PluginConfig{
			&PluginConfig{
				Name:    "transpile",
				Command: "echo",
				Args:    "-n {{fileContent}}1",
			},
			&PluginConfig{
				Name:    "template",
				Command: "echo",
				Args:    "-n {{fileContent}}2",
				PipeTo:  "transpile",
			},
		}

		plugins := NewPlugins(pcs)

		Convey("Names should be correct", func() {
			p := plugins.Get("transpile")
			So(p.Name, ShouldEqual, "transpile")
		})

		Convey("Tranformers should be correct", func() {
			p := plugins.Get("transpile")
			p.InC <- &File{Name: "foo.js", Content: "hello"}
			res := <-p.OutC

			So(res.Name, ShouldEqual, "foo.js")
			So(res.Content, ShouldEqual, "hello1")
		})

		Convey("Plugins that pipe to another plugin have their OutC set", func() {
			p := plugins.Get("template")
			p.InC <- &File{Name: "foo.js", Content: "hello"}
			res := <-plugins.Get("transpile").OutC

			So(res.Name, ShouldEqual, "foo.js")
			So(res.Content, ShouldEqual, "hello21")
		})
	})
}
