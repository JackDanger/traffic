package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"

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

	har, err := parser.HarFrom(form.Har.Source)
	if err != nil {
		fail(err, w)
		return
	}

	db, err := persistence.NewDb()
	if err != nil {
		fail(err, w)
		return
	}

	archive := persistence.MakeArchive(har)
	_, err = db.Create(archive)

	w.Write(archive.AsJSON())
}

// StartHar begins 1 or more runners of a specific HAR file identified by name
// (for now, eventually it'll be by token from the db)
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
