package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"

	"github.com/JackDanger/traffic/parser"
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
		fail(err, w)
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
		fail(err, w)
		return
	}

	content := []parser.HarWrapper{
		parser.HarWrapper{Har: *har},
	}
	contentJSON, err := json.Marshal(content)
	if err != nil {
		fail(err, w)
		return
	}

	w.WriteHeader(200)
	w.Write(contentJSON)
}

// CreateHar stores a new HAR
func CreateHar(w http.ResponseWriter, r *http.Request) {
	// Extracts form values into a url.Values (map[string][]string) instance.
	// Note that the popular nesting convention doesn't work natively in Go:
	//   "?key[subkey]=x" -> map[string][]string{"key[subkey]": "x"}
	// But if you use dot notation like the following then Gorilla's schema can
	// extract nested objects:
	//   "?key.subkey=x" -> map[string][]string{"key.subkey": "x"}
	err := r.ParseForm()
	if err != nil {
		fail(err, w)
		return
	}

	type H struct {
		Name   string `json:"name",schema:"name"`
		Source string `json:"source",schema:"source"`
	}
	type Form struct {
		Har H `json:"form",schema:"har"`
	}
	form := &Form{}
	decoder := schema.NewDecoder()
	// r.PostForm is a map of our POST form values
	err = decoder.Decode(form, r.PostForm)
	if err != nil {
		fail(err, w)
		return
	}

	contentJSON, err := json.Marshal(form)
	if err != nil {
		fail(err, w)
		return
	}

	content := "submitted web form of length: " + strconv.Itoa(int(r.ContentLength)) + " bytes"
	content += "\n"
	content += string(contentJSON)
	w.Write([]byte(content))
}

// StartHar begins 1 or more runners of a specific HAR file identified by name
// (for now, eventually it'll be by token from the db)
func StartHar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// TODO: read from the db
	har, err := parser.HarFrom(currentDir() + "/../fixtures/browse-two-github-users.har")

	if err != nil {
		fail(err, w)
		return
	}

	content := []parser.HarWrapper{
		parser.HarWrapper{Har: *har},
	}
	contentJSON, err := json.Marshal(content)
	if err != nil {
		fail(err, w)
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

func fail(err error, w http.ResponseWriter) {
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}
