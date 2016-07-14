package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"runtime"

	"github.com/JackDanger/traffic/parser"
	"github.com/gorilla/mux"
)

// NewServer returns an instance of http.Server ready to listen on the given
// port. Calling ListenAndServe on it is a blocking call and should likely be
// in a goroutine somewhere:
//   server.NewServer("8080").ListenAndServe()
func NewServer(port string) *http.Server {

	r := mux.NewRouter()

	r.HandleFunc("/", Index).Methods("GET")
	r.HandleFunc("/hars", ListHars).Methods("GET")
	r.HandleFunc("/hars", CreateHar).Methods("POST")
	r.HandleFunc("/start", StartHar).Methods("POST")

	handler := http.NewServeMux()
	handler.Handle("/", r)

	return &http.Server{
		Addr:    "127.0.0.1:" + port,
		Handler: handler,
	}
}

// Index shows the home page
func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	content, err := webFile("index.html")

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write(content)
}

// ListHars retrieves all HAR files
func ListHars(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// TODO: read from the db
	har, err := parser.HarFrom(currentDir() + "/../fixtures/browse-two-github-users.har")

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	content := []parser.HarWrapper{
		parser.HarWrapper{Har: *har},
	}
	contentJSON, err := json.Marshal(content)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(200)
	w.Write(contentJSON)
}

// CreateHar stores a new HAR
func CreateHar(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("sure, let's get around to persistence eventually.\n"))
}

// StartHar begins 1 or more runners of a specific HAR file identified by name
// (for now, eventually it'll be by token from the db)
func StartHar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// TODO: read from the db
	har, err := parser.HarFrom(currentDir() + "/../fixtures/browse-two-github-users.har")

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	content := []parser.HarWrapper{
		parser.HarWrapper{Har: *har},
	}
	contentJSON, err := json.Marshal(content)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(200)
	w.Write(contentJSON)
}

// Points to this file's parent directory
func currentDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return filepath.Base(path.Dir(filename))
}

// Extracts the contents of a provided file from "server/_site/:filename"
func webFile(filename string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(currentDir() + "/_site/" + filename)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
