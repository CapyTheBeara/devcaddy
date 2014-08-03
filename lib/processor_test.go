package lib

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func createFile() *File {
	return &File{Name: "foo.js", Content: "hello"}
}

func TestProcessor(t *testing.T) {
	inputFile := createFile()

	Convey("Given a Processor with a transformer function", t, func() {

		fn := func(f *File) *File {
			return &File{Name: f.Name + "1", Content: f.Content + "!"}
		}

		Convey("It can manipulate it's input", func() {
			p := NewProcessor(&ProcessorConfig{}, fn)
			p.InC <- inputFile
			res := <-p.OutC

			So(res.Name, ShouldEqual, "foo.js1")
			So(res.Content, ShouldEqual, "hello!")
		})

		Convey("It can be configured to not send it's transformer output", func() {
			p := NewProcessor(&ProcessorConfig{NoOutput: true}, fn)
			p.InC <- inputFile
			res := <-p.OutC

			So(res, ShouldBeNil)
		})
	})
}

func TestCommandProcessor(t *testing.T) {
	inputFile := createFile()

	Convey("Given a CommandProcessor", t, func() {
		Convey("An input file's name and content is sent as params", func() {
			p := NewCommandProcessor(&ProcessorConfig{
				Command: "echo",
				Args:    "-n",
			})
			p.InC <- inputFile
			res := <-p.OutC

			So(res.Name, ShouldEqual, "foo.js")
			So(res.Content, ShouldEqual, "foo.js hello")
		})

		Convey("Command error is added to the output", func() {
			p := NewCommandProcessor(&ProcessorConfig{
				Command: "a",
			})
			p.InC <- inputFile
			res := <-p.OutC

			So(res.Name, ShouldEqual, "foo.js")
			So(res.Content, ShouldEqual, "")
			So(res.Error, ShouldNotBeNil)
		})

		Convey("Command can change the output file's name", func() {
			p := NewCommandProcessor(&ProcessorConfig{
				Command: "echo",
				Args:    "-n __SERVER_OUTPUT_PATH__=",
			})
			p.InC <- inputFile
			res := <-p.OutC

			So(res.Name, ShouldEqual, " foo.js hello")
			So(res.Content, ShouldEqual, "")
		})
	})
}
