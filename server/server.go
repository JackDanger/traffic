package server

import (
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"runtime"
)

// this value will be initialized by `webFile` to hold the directory containing
// thie source file.
var currentDir string

// Route is a very small DSL for defining route/handler pairs
type Route struct {
	Path    string
	Handler http.HandlerFunc
}

var routes = []Route{
	{
		"/",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			content, err := webFile("index.html")

			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(200)
			w.Write(content)
		},
	},
}

// NewServer returns an instance of http.Server ready to listen on the given
// port. Calling ListenAndServe on it is a blocking call and should likely be
// in a goroutine somewhere:
//   server.NewServer("8080").ListenAndServe()
func NewServer(port string) *http.Server {
	server := http.Server{
		Addr: "127.0.0.1:" + port,
	}
	for _, route := range routes {
		http.HandleFunc(route.Path, route.Handler)
	}
	return &server
}

// Extracts the contents of a provided file from "server/_site/:filename"
func webFile(filename string) ([]byte, error) {
	if currentDir == "" {
		_, filename, _, _ := runtime.Caller(1)
		currentDir = filepath.Base(path.Dir(filename))
	}
	bytes, err := ioutil.ReadFile(currentDir + "/_site/" + filename)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
