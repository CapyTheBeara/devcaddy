package lib

import (
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/fsnotify.v0"
)

func TestFileWatcher(t *testing.T) {
	SetDefaultFailureMode(FailureContinues)

	Convey("Given a file watcher", t, func() {
		dir := "../tmp1"
		removeTestDir(t, dir)
		makeTestDir(t, dir+"/foo")
		makeTestDir(t, dir+"/bar")
		makeTestFile(t, dir, "foo/index.js", "var foo;\n", 0)
		makeTestFile(t, dir, "bar/main.js", "1", 0)
		makeTestFile(t, dir, "foo/nope.foo", "nope", 0)

		Convey("If no plugins are given", func() {
			c := WatcherConfig{
				Dir:   dir,
				Files: []string{"foo/index.js", "bar/main.js"},
			}

			config := Config{Plugins: &Plugins{}}
			out := make(chan *File)
			w := NewWatcher(dir, out, &c, &config)

			Convey("GetAllFiles passes the correct files unmodified if no plugin given", func() {
				defer removeTestDir(t, dir)

				doneC := make(chan bool)
				go func() {
					f1 := <-out
					So(f1.Name, ShouldEqual, "../tmp1/foo/index.js")
					So(f1.Content, ShouldEqual, "var foo;\n")

					f2 := <-out
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

				f1 := <-out
				So(f1.Name, ShouldEqual, "../tmp1/foo/index.js")
				So(f1.Content, ShouldEqual, "var foo;\ns")

				makeTestFile(t, dir, "app.hbs", "1", 20)

				select {
				case <-out:
					So("Failed - Wrong file received", ShouldBeNil)
				default:
					So("Passed - wrong file not received", ShouldNotBeBlank)
				}
			})
		})

		Convey("If a plugin is given, the file is modified by the plugin", func() {
			defer removeTestDir(t, dir)

			c := WatcherConfig{
				Dir:         dir,
				Files:       []string{"foo/index.js", "bar/main.js"},
				PluginNames: []string{"zzz"},
			}

			p := NewPlugin(&PluginConfig{Name: "zzz"}, func(f *File) *File {
				return &File{Name: "zzz"}
			})
			config := Config{Plugins: &Plugins{content: map[string]*Plugin{"zzz": p}}}
			out := make(chan *File)
			w := NewWatcher(dir, out, &c, &config)

			<-w.Ready()
			updateTestFile(t, "../tmp1/foo/index.js", "s")

			f1 := <-out
			So(f1.Name, ShouldEqual, "zzz")

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
		makeTestFile(t, dir, "bar/main.js", "1", 0)
		makeTestFile(t, dir, "nope.foo", "nope", 0)

		c := WatcherConfig{
			Dir:         dir,
			Ext:         "js",
			PluginNames: []string{"transpile-js"},
		}

		p := NewCommandPlugin(&PluginConfig{
			Name:    "transpile-js",
			Command: "echo",
			Args:    "-n",
		})

		config := Config{Plugins: &Plugins{content: map[string]*Plugin{"transpile-js": p}}}
		out := make(chan *File)
		w := NewWatcher("", out, &c, &config)

		Convey("GetAllFiles passes the correct files modified by Plugins", func() {
			defer removeTestDir(t, dir)

			doneC := make(chan bool)
			go func() {
				f := <-out
				So(f.Name, ShouldEqual, "../tmp2/bar/main.js")
				So(f.Content, ShouldEqual, "../tmp2/bar/main.js 1")

				f = <-out
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

			f := <-out
			So(f.Name, ShouldEqual, "../tmp2/foo.js")
			So(f.Content, ShouldEqual, "../tmp2/foo.js var foo;\ns")

			// test subdir is watched
			updateTestFile(t, "../tmp2/bar/main.js", "d")

			f = <-out
			So(f.Name, ShouldEqual, "../tmp2/bar/main.js")
			So(f.Content, ShouldEqual, "../tmp2/bar/main.js 1d")

			// test creating new subdir is watched
			makeTestDir(t, dir+"/baz", 20)
			makeTestFile(t, dir+"/baz", "baz.js", "", 0)

			f = <-out
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

			f := <-out
			So(f.Name, ShouldEqual, "../tmp2/foo.js")

			f = <-out
			So(f.Name, ShouldEqual, "../tmp2/bar/main.js")

			time.Sleep(time.Millisecond * 20)

			select {
			case <-out:
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

			f := <-out
			So(f.Name, ShouldEqual, "../tmp2/foo.js")

			time.Sleep(time.Millisecond * 20)

			select {
			case <-out:
				So("Passes - second event is processed", ShouldNotBeBlank)
			default:
				So("Fails - shouldn't suppress second event", ShouldBeNil)
			}
		})

		Convey("An error event is passed through as a file", func() {
			defer removeTestDir(t, dir)
			err := errors.New("")

			w.fsWatcher().Errors <- err
			f := <-out
			So(f.Name, ShouldEqual, "watcher error")
			So(f.Content, ShouldEqual, "")
			So(f.Op, ShouldEqual, ERROR)
			So(f.Error, ShouldEqual, err)
		})
	})
}

func TestGroupAllWatcher(t *testing.T) {
	SetDefaultFailureMode(FailureContinues)

	Convey("Given a group all watcher", t, func() {
		dir := "../tmp3"
		removeTestDir(t, dir)
		makeTestDir(t, dir+"/styles/partials")
		makeTestDir(t, dir+"/styles/vendor")
		makeTestFile(t, dir, "styles/app.scss", "1", 0)
		makeTestFile(t, dir, "styles/partials/_foo.scss", "2", 0)
		makeTestFile(t, dir, "styles/vendor/bar.scss", "3", 0)
		makeTestFile(t, dir, "nope.foo", "nope", 0)

		c := WatcherConfig{
			Dir:         dir,
			Ext:         "scss",
			GroupAll:    true,
			PluginNames: []string{"sass"},
		}

		p := NewCommandPlugin(&PluginConfig{
			Name:    "sass",
			Command: "echo",
			Args:    "-n {{fileContent}}",
		})

		config := Config{Plugins: &Plugins{content: map[string]*Plugin{"sass": p}}}
		out := make(chan *File)
		w := NewWatcher("", out, &c, &config)

		Convey("GetAllFiles passes to the plugins the content of all of the files being watched that have already been processed", func() {
			defer removeTestDir(t, dir)

			doneC := make(chan bool)
			go func() {
				f := <-out
				So(f.Name, ShouldEqual, "../tmp3/styles/app.scss")
				So(f.Content, ShouldEqual, "1")

				f = <-out
				So(f.Name, ShouldEqual, "../tmp3/styles/partials/_foo.scss")
				So(f.Content, ShouldEqual, "1\n2")

				f = <-out

				So(f.Name, ShouldEqual, "../tmp3/styles/vendor/bar.scss")
				So(f.Content, ShouldEqual, "1\n2\n3")

				time.Sleep(time.Millisecond * 20)

				select {
				case <-out:
					So("Fail - Shouldn't get here", ShouldBeNil)
				default:
					So("Pass - Inappropriate file is not processed", ShouldNotBeBlank)
				}
				doneC <- true
			}()

			w.GetAllFiles()
			<-doneC
		})
	})
}

func TestProxyWatcher(t *testing.T) {
	SetDefaultFailureMode(FailureContinues)

	Convey("Given a proxy watcher", t, func() {
		dir := "../tmp4"
		removeTestDir(t, dir)
		makeTestDir(t, dir+"/styles/partials")
		makeTestFile(t, dir, "styles/app.scss", "1", 0)
		makeTestFile(t, dir, "styles/partials/_foo.scss", "2", 0)
		makeTestFile(t, dir, "styles/nope.foo", "nope", 0)

		c := WatcherConfig{
			Dir:         "styles",
			Ext:         "scss",
			Proxy:       "styles/app.scss",
			PluginNames: []string{"sass"},
		}

		p := NewCommandPlugin(&PluginConfig{
			Name:    "sass",
			Command: "echo",
			Args:    "-n {{fileContent}}",
		})

		config := Config{Plugins: &Plugins{content: map[string]*Plugin{"sass": p}}}
		out := make(chan *File)
		w := NewWatcher(dir, out, &c, &config)

		Convey("GetAllFiles only passes proxy once", func() {
			defer removeTestDir(t, dir)

			<-w.Ready()

			doneC := make(chan bool)
			go func() {
				f1 := <-out
				So(f1.Name, ShouldEqual, "../tmp4/styles/app.scss")
				So(f1.Content, ShouldEqual, "1")

				time.Sleep(time.Millisecond * 20)

				select {
				case <-out:
					So("Fail - Shouldn't get here", ShouldBeNil)
				default:

					So("Pass - Proxy is not sent again", ShouldNotBeBlank)
				}
				doneC <- true
			}()

			w.GetAllFiles()
			<-doneC
		})

		Convey("Instead of a watched file being sent to the plugin, the proxy is sent", func() {
			defer removeTestDir(t, dir)

			<-w.Ready()
			updateTestFile(t, "../tmp4/styles/partials/_foo.scss", "s")

			f := <-out
			So(f.Name, ShouldEqual, "../tmp4/styles/app.scss")
			So(f.Content, ShouldEqual, "1")
		})
	})
}

func TestNewWatchers(t *testing.T) {
	Convey("Given a Config", t, func() {
		c := NewConfig([]byte(`
    		{
    		    "root": "../mockapp",
    		    "watch": [
		        {
		            "dir": "app/templates",
		            "ext": "hbs"
		        }
		    ],
		    "files": [
		        {
		            "name": "app.js",
		            "dir": "app",
		            "ext": "js"
		        }
		     ]
	    }`))

		out := make(chan *File)
		ws := NewWatchers(c, out)

		Convey("It creates watchers from the 'watch' setting", func() {
			w := ws.Get("app/templates:hbs")
			So(w.Name(), ShouldEqual, "app/templates:hbs")
		})

		Convey("It creates watchers from the 'files' setting", func() {
			w := ws.Get("app.js")
			So(w.Name(), ShouldEqual, "app.js")
		})

	})
}
