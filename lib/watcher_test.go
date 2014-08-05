package lib

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/fsnotify.v0"
)

func TestFileWatcher(t *testing.T) {
	SetDefaultFailureMode(FailureContinues)

	Convey("Given a file watcher", t, func() {
		dir := "../tmp1"
		makeTestDir(t, dir+"/foo")
		makeTestDir(t, dir+"/bar")
		makeTestFile(t, dir+"/foo", "index.js", "var foo;\n", 0)
		makeTestFile(t, dir+"/bar", "main.js", "1", 0)
		makeTestFile(t, dir+"/foo", "nope.foo", "nope", 0)

		c := WatcherConfig{
			Dir:   dir,
			Files: []string{"foo/index.js", "bar/main.js"},
		}
		config := Config{Processors: []*Processor{}}
		w := NewWatcher(dir, make(chan *File), &c, &config)

		Convey("GetAllFiles passes the correct files unmodified", func() {
			defer removeTestDir(t, dir)

			doneC := make(chan bool)
			go func() {
				f1 := <-w.OutChan()
				So(f1.Name, ShouldEqual, "../tmp1/foo/index.js")
				So(f1.Content, ShouldEqual, "var foo;\n")

				f2 := <-w.OutChan()
				So(f2.Name, ShouldEqual, "../tmp1/bar/main.js")
				doneC <- true
			}()

			w.GetAllFiles()
			<-doneC
		})

		Convey("A file change on a relavent file is detected", func() {
			defer removeTestDir(t, dir)

			<-w.Ready()
			updateTestFile(t, "../tmp1/foo/index.js", "s")

			f1 := <-w.OutChan()
			So(f1.Name, ShouldEqual, "../tmp1/foo/index.js")
			So(f1.Content, ShouldEqual, "var foo;\ns")

			makeTestFile(t, dir, "app.hbs", "1", 20)

			select {
			case <-w.OutChan():
				So("Failed - Wrong file received", ShouldBeNil)
			default:
				So("Passed - wrong file not received", ShouldNotBeBlank)
			}
		})
	})
}

func TestDirWatcher(t *testing.T) {
	SetDefaultFailureMode(FailureContinues)

	Convey("Given a dir watcher", t, func() {
		dir := "../tmp2"
		removeTestDir(t, dir)
		makeTestDir(t, dir+"/bar")
		makeTestFile(t, dir, "foo.js", "var foo;\n", 0)
		makeTestFile(t, dir+"/bar", "main.js", "1", 0)
		makeTestFile(t, dir, "nope.foo", "nope", 0)

		c := WatcherConfig{
			Dir:     dir,
			Ext:     "js",
			Plugins: []string{"transpile-js"},
		}

		p := NewCommandProcessor(&ProcessorConfig{
			Name:    "transpile-js",
			Command: "echo",
			Args:    "-n",
		})

		config := Config{Processors: []*Processor{p}}
		w := NewWatcher("", make(chan *File), &c, &config)

		Convey("GetAllFiles passes the correct files modified by Processors", func() {
			defer removeTestDir(t, dir)

			doneC := make(chan bool)
			go func() {
				f := <-w.OutChan()
				So(f.Name, ShouldEqual, "../tmp2/bar/main.js")
				So(f.Content, ShouldEqual, "../tmp2/bar/main.js 1")

				f = <-w.OutChan()
				So(f.Name, ShouldEqual, "../tmp2/foo.js")
				So(f.Content, ShouldEqual, "../tmp2/foo.js var foo;\n")
				doneC <- true
			}()

			w.GetAllFiles()
			<-doneC
		})

		Convey("A file change on a relavent file is detected", func() {
			defer removeTestDir(t, dir)

			<-w.Ready()
			updateTestFile(t, "../tmp2/foo.js", "s")

			f := <-w.OutChan()
			So(f.Name, ShouldEqual, "../tmp2/foo.js")
			So(f.Content, ShouldEqual, "../tmp2/foo.js var foo;\ns")

			// test subdir is watched
			updateTestFile(t, "../tmp2/bar/main.js", "d")

			f = <-w.OutChan()
			So(f.Name, ShouldEqual, "../tmp2/bar/main.js")
			So(f.Content, ShouldEqual, "../tmp2/bar/main.js 1d")

			// test creating new subdir is watched
			makeTestDir(t, dir+"/baz", 20)
			makeTestFile(t, dir+"/baz", "baz.js", "", 0)

			f = <-w.OutChan()
			So(f.Name, ShouldEqual, "../tmp2/baz/baz.js")
			So(f.Content, ShouldEqual, "../tmp2/baz/baz.js ")
		})

		Convey("Multple identical events within 100ms are considered as one", func() {
			defer removeTestDir(t, dir)

			evt := fsnotify.Event{Name: "../tmp2/foo.js", Op: fsnotify.Create}
			evt2 := fsnotify.Event{Name: "../tmp2/bar/main.js", Op: fsnotify.Create}

			w.fsWatcher().Events <- evt
			w.fsWatcher().Events <- evt2
			w.fsWatcher().Events <- evt

			f := <-w.OutChan()
			So(f.Name, ShouldEqual, "../tmp2/foo.js")

			f = <-w.OutChan()
			So(f.Name, ShouldEqual, "../tmp2/bar/main.js")

			select {
			case <-w.OutChan():
				So("Fails - shouldn't process the second event", ShouldBeNil)
			default:
				So("Passes - second event is not processed", ShouldNotBeBlank)
			}
		})

		Convey("Multiple identical events >100ms apart are considered separate", func() {
			defer removeTestDir(t, dir)

			evt := fsnotify.Event{Name: "../tmp2/foo.js", Op: fsnotify.Create}
			w.fsWatcher().Events <- evt

			time.Sleep(time.Millisecond * 100)
			w.fsWatcher().Events <- evt

			f := <-w.OutChan()
			So(f.Name, ShouldEqual, "../tmp2/foo.js")
			time.Sleep(time.Millisecond * 20)

			select {
			case <-w.OutChan():
				So("Passes - second event is processed", ShouldNotBeBlank)
			default:
				So("Fails - shouldn't suppress second event", ShouldBeNil)
			}
		})
	})
}
