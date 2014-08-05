package lib

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"

	"code.google.com/p/go.net/websocket"
	. "github.com/smartystreets/goconvey/convey"
)

func newTestWR(t *testing.T, url string) (*httptest.ResponseRecorder, *http.Request) {
	req, err := http.NewRequest("", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewRecorder(), req
}

func TestHtmls(t *testing.T) {
	Convey("Given a server with a store", t, func() {
		store := Store{}
		store.Put("index.html", "index!")
		store.Put("tests.html", "tests!")

		s := NewServer(&store)

		Convey("Html serves .html files from the store", func() {
			w, r := newTestWR(t, "/")
			s.Html(w, r)
			So(w.Body.String(), ShouldEqual, "index!")

			w, r = newTestWR(t, "/index.html")
			s.Html(w, r)
			So(w.Body.String(), ShouldEqual, "index!")

			w, r = newTestWR(t, "/tests.html")
			s.Html(w, r)
			So(w.Body.String(), ShouldEqual, "tests!")

			w, r = newTestWR(t, "/tests")
			s.Html(w, r)
			So(w.Body.String(), ShouldEqual, "tests!")

			w, r = newTestWR(t, "/howdy")
			s.Html(w, r)
			So(w.Body.String(), ShouldContainSubstring, "howdy.html was not found")
		})

		Convey("Html can prepend content to 'index.html'", func() {
			s.PrependIndex = "foo"
			w, r := newTestWR(t, "/")
			s.Html(w, r)
			So(w.Body.String(), ShouldEqual, "foo\nindex!")
		})

		Convey("Html sends other requests to a reverse proxy if set", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("proxy here"))
			}))
			defer ts.Close()

			backendURL, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatal(err)
			}

			s.Proxy = httputil.NewSingleHostReverseProxy(backendURL)

			w, r := newTestWR(t, "/")
			s.Html(w, r)
			So(w.Body.String(), ShouldEqual, "index!")

			w, r = newTestWR(t, "/howdy")
			s.Html(w, r)
			So(w.Body.String(), ShouldEqual, "proxy here")
		})
	})
}

func TestAssets(t *testing.T) {
	Convey("Given a server with a store", t, func() {
		store := Store{}
		store.Put("app.js", "app js")
		store.Put("app.css", "app css")

		s := NewServer(&store)

		Convey("Assets sends corresponding files from the store", func() {
			w, r := newTestWR(t, "/assets/app.js")
			s.Assets(w, r)
			So(w.Body.String(), ShouldEqual, "app js")
			So(w.Header().Get("Content Type"), ShouldEqual, "text/javascript")

			w, r = newTestWR(t, "/assets/app.css")
			s.Assets(w, r)
			So(w.Body.String(), ShouldEqual, "app css")
			So(w.Header().Get("Content Type"), ShouldEqual, "text/css")
		})

		Convey("Asset root can be changed", func() {
			s.AssetRoot = "static"
			w, r := newTestWR(t, "/static/app.js")
			s.Assets(w, r)

			So(w.Body.String(), ShouldEqual, "app js")
			So(w.Header().Get("Content Type"), ShouldEqual, "text/javascript")
		})
	})
}

func TestWebsocket(t *testing.T) {
	Convey("Given a test server", t, func() {
		store := Store{DidUpdate: make(chan string)}
		s := NewServer(&store)
		ts := httptest.NewServer(s.Websocket())
		addr := ts.Listener.Addr().String()

		Convey("WS clients should receive 'RELOAD' message when the store is updated", func() {
			defer ts.Close()

			ws, err := websocket.Dial("ws://"+addr, "", "http://localhost/")
			if err != nil {
				t.Fatal(err)
			}

			ws2, err := websocket.Dial("ws://"+addr, "", "http://localhost/")
			if err != nil {
				t.Fatal(err)
			}

			So(len(s.ReloadChans), ShouldEqual, 2)
			store.Put("fooz", "bar")

			var msg WSMessage
			if err = websocket.JSON.Receive(ws, &msg); err != nil {
				t.Fatal(err)
			}

			So(msg.Message, ShouldEqual, "RELOAD")
			So(msg.File, ShouldEqual, "fooz")

			var msg2 WSMessage
			if err = websocket.JSON.Receive(ws2, &msg2); err != nil {
				t.Fatal(err)
			}

			So(msg2.Message, ShouldEqual, "RELOAD")
			So(msg2.File, ShouldEqual, "fooz")
			So(len(s.ReloadChans), ShouldEqual, 0)

		})
	})
}
