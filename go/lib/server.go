package lib

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"

	"code.google.com/p/go.net/websocket"
)

var assetTypes = map[string]string{
	".js":  "text/javascript",
	".css": "text/css",
}

type WSMessage struct {
	Message string
	File    string
}

func NewServer(store *Store) *Server {
	server := Server{
		Store:       store,
		ReloadChans: make(map[chan string]bool),
	}

	go server.listenForStoreUpdates()
	return &server
}

// args = [port, proxy, assetRoot]
func StartServer(store *Store, port, prox, assetRoot string) *Server {
	s := NewServer(store)

	if prox != "" {
		u, err := url.Parse(prox)
		if err != nil {
			log.Fatal("Error parsing proxy URL", err)
		}
		proxy := httputil.NewSingleHostReverseProxy(u)
		s.Proxy = proxy
	}

	s.AssetRoot = assetRoot
	s.PrependIndex = reloadScript(port)

	http.Handle("/reload/", s.Websocket())

	http.HandleFunc("/assets/", s.Assets)
	http.HandleFunc("/", s.Html)

	log.Fatal(http.ListenAndServe(":"+port, nil))
	return s
}

type Server struct {
	Store        *Store
	Proxy        *httputil.ReverseProxy
	PrependIndex string
	AssetRoot    string
	ReloadChans  map[chan string]bool
}

func (s *Server) Html(w http.ResponseWriter, r *http.Request) {
	file := ""
	path := r.URL.Path[1:]
	base := filepath.Base(path)

	if path == "" {
		file = "index.html"
	} else if filepath.Ext(base) == ".html" {
		file = path
	} else {
		file = path + ".html"
	}

	content := s.Store.Get(file)
	if content == "" {
		if s.Proxy != nil {
			s.Proxy.ServeHTTP(w, r)
			return
		} else {
			content = file + " was not found in your store. Make sure to define it in your config file or specify a proxy."
		}
	}

	prepend := ""
	if s.PrependIndex != "" {
		prepend = s.PrependIndex + "\n"
	}

	fmt.Fprint(w, prepend+content)
}

func (s *Server) Assets(w http.ResponseWriter, r *http.Request) {
	root := "assets"
	if s.AssetRoot != "" {
		root = s.AssetRoot
	}

	name := r.URL.Path[len("/"+root+"/"):]
	file := s.Store.Get(name)

	w.Header().Set("Content Type", assetTypes[filepath.Ext(name)])
	fmt.Fprint(w, file)
}

func (s *Server) Websocket() http.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		reload := make(chan string)
		s.ReloadChans[reload] = true

		name := <-reload
		msg := WSMessage{Message: "RELOAD", File: name}

		err := websocket.JSON.Send(ws, msg)
		if err != nil {
			log.Println("[ws error]", err)
		}

		delete(s.ReloadChans, reload)
	})
}

func (s *Server) listenForStoreUpdates() {
	for {
		name := <-s.Store.DidUpdate
		Plog.PrintC("server", "updated: "+name)
		for ch, _ := range s.ReloadChans {
			ch <- name
		}
	}
}
