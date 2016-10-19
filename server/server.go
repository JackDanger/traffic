package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/JackDanger/traffic/parser"
	"github.com/JackDanger/traffic/persistence"
	"github.com/JackDanger/traffic/util"
)

// NewServer returns an instance of http.Server ready to listen on the given
// port. Calling ListenAndServe on it is a blocking call:
//   server.NewServer("8080").ListenAndServe()
func NewServer(port string) *http.Server {

	r := mux.NewRouter()

	r.HandleFunc("/", Index).Methods("GET")
	r.HandleFunc("/javascript.js", Javascript).Methods("GET")
	r.HandleFunc("/archives", ListHars).Methods("GET")
	r.HandleFunc("/archives", CreateHar).Methods("POST")
	r.HandleFunc("/archives/{id}", UpdateHar).Methods("PUT")
	r.HandleFunc("/archives/{id}", DeleteHar).Methods("DELETE")
	r.HandleFunc("/start", StartHar).Methods("POST")

	handler := newLoggedMux()
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

// Javascript turns JSX into JS and renders is
func Javascript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	content, err := webFile("javascript.js")

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
	db, err := persistence.NewDb()
	if err != nil {
		fail(err, w)
		return
	}

	archives, err := db.ListArchives()
	if err != nil {
		fail(err, w)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("["))
	for i := 0; i < len(archives)-1; i++ {
		w.Write(append(archives[i].AsJSON(), byte(',')))
	}
	// And don't add a trailing comma on the last one
	if len(archives) > 0 {
		w.Write(archives[len(archives)-1].AsJSON())
	}
	w.Write([]byte("]"))
}

// HarParam is the archive data nested inside Form
type HarParam struct {
	Name        string `json:"name",schema:"name"`
	Description string `json:"description",schema:"description"`
	Source      string `json:"source",schema:"source"`
}

// CreateHar stores a new HAR
func CreateHar(w http.ResponseWriter, r *http.Request) {
	// Extracts form values into a url.Values (map[string][]string) instance.
	// Note that the popular nesting convention doesn't work natively in Go:
	//   "?key[subkey]=x" -> map[string][]string{"key[subkey]": "x"}
	// But if you use dot notation like the following then Gorilla's schema can
	// extract nested objects:
	//   "?key.subkey=x" -> map[string][]string{"key": map[string]string{"subkey": "x"}}
	err := r.ParseForm()
	if err != nil {
		fail(err, w)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fail(err, w)
		return
	}

	harParam := &HarParam{}
	err = json.Unmarshal(body, harParam)
	if err != nil {
		fail(err, w)
		return
	}

	har, err := parser.HarFrom(harParam.Source)
	if err != nil {
		fail(err, w)
		return
	}

	db, err := persistence.NewDb()
	if err != nil {
		fail(err, w)
		return
	}

	archive := persistence.MakeArchive(harParam.Name, harParam.Description, har)
	_, err = db.Create(archive)

	w.Write(archive.AsJSON())
}

// UpdateHar stores a new HAR
func UpdateHar(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if params["id"] == "" {
		w.WriteHeader(404)
		w.Write([]byte(`{"success": "nope"}`))
	}
	w.WriteHeader(200)
	w.Write([]byte(`{"success": "sure"}`))
}

// DeleteHar stores a new HAR
func DeleteHar(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(`{"success": "sure"}`))
}

// StartHar begins 1 or more runners of a specific HAR file identified by name
// (for now, eventually it'll be by id from the db)
func StartHar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// TODO: read from the db
	har, err := parser.HarFrom(util.Root() + "fixtures/browse-two-github-users.har")

	if err != nil {
		fail(err, w)
		return
	}

	content := []parser.HarWrapper{
		parser.HarWrapper{Har: har},
	}
	contentJSON, err := json.Marshal(content)
	if err != nil {
		fail(err, w)
		return
	}

	w.WriteHeader(200)
	w.Write(contentJSON)
}

// LoggedMux is a wrapper around http.ServeMux that logs all requests to STDERR
type LoggedMux struct {
	*http.ServeMux
	log *log.Logger
}

func newLoggedMux() *LoggedMux {
	var mux = &LoggedMux{}
	mux.ServeMux = http.NewServeMux()
	mux.log = log.New(os.Stderr, "", log.LstdFlags)
	return mux
}

// ServeHTTP delegates to the internal ServeMux and then logs
func (mux *LoggedMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux.ServeMux.ServeHTTP(w, r)
	mux.log.Printf("%s %s", r.Method, r.URL.Path)
}

// Extracts the contents of a provided file from "server/_site/:filename"
func webFile(filename string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(util.Root() + "server/_site/" + filename)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func fail(err error, w http.ResponseWriter) {
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}
