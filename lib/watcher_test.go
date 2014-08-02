package lib

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFileWatcher(t *testing.T) {
	Convey("Given a file watcher", t, func() {
		c := WatcherConfig{
			Dir:   "vendor",
			Files: []string{"bar/index.js", "baz/main.js"},
		}
		config := Config{Processors: []*Processor{}}
		w := NewWatcher("../mockapp", make(chan *File), &c, &config)

		Convey("GetAllFiles passes the correct files unmodified", func() {
			doneC := make(chan bool)

			go func() {
				f1 := <-w.OutChan()
				So(f1.Name, ShouldEqual, "../mockapp/vendor/bar/index.js")
				So(f1.Content, ShouldEqual, "var bar;\n")

				f2 := <-w.OutChan()
				So(f2.Name, ShouldEqual, "../mockapp/vendor/baz/main.js")
				doneC <- true
			}()

			w.GetAllFiles()
			<-doneC
		})
	})
}

func TestDirWatcher(t *testing.T) {
	Convey("Given a dir watcher", t, func() {
		c := WatcherConfig{
			Dir:     "app",
			Ext:     "js",
			Plugins: []string{"transpile-js"},
		}

		p := NewCommandProcessor("echo", []string{"-n"})
		p.Name = "transpile-js"

		config := Config{Processors: []*Processor{p}}
		w := NewWatcher("../mockapp", make(chan *File), &c, &config)

		Convey("GetAllFiles passes the correct files modified by Processors", func() {
			doneC := make(chan bool)

			go func() {
				f1 := <-w.OutChan()
				So(f1.Name, ShouldEqual, "../mockapp/app/foo.js")
				So(f1.Content, ShouldEqual, "../mockapp/app/foo.js var foo;\n")

				f2 := <-w.OutChan()
				So(f2.Name, ShouldEqual, "../mockapp/app/main.js")
				doneC <- true
			}()

			w.GetAllFiles()
			<-doneC
		})
	})
}
