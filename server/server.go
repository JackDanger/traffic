package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/JackDanger/traffic/parser"
	"github.com/JackDanger/traffic/persistence"
	"github.com/JackDanger/traffic/util"
)

// Store a single database connection instance for the whole server. Is this a
// terrible idea?
var db *persistence.DB

// NewServer returns an instance of http.Server ready to listen on the given
// port. Calling ListenAndServe on it is a blocking call:
//   server.NewServer("8080").ListenAndServe()
func NewServer(port string) (*http.Server, error) {

	r := mux.NewRouter()

	r.HandleFunc("/", Index).Methods("GET")
	r.HandleFunc("/javascript.js", Javascript).Methods("GET")
	r.HandleFunc("/archives", ListArchives).Methods("GET")
	r.HandleFunc("/archives", CreateArchive).Methods("POST")
	r.HandleFunc("/archives/{id}", UpdateArchive).Methods("PUT")
	r.HandleFunc("/archives/{id}", DeleteArchive).Methods("DELETE")
	r.HandleFunc("/start", StartHar).Methods("POST")

	handler := newLoggedMux()
	handler.Handle("/", r)

	var err error
	// Set the global `db` var WTF please open a pull request to fix how I'm
	// doing this.
	db, err = persistence.NewDb()
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:    "127.0.0.1:" + port,
		Handler: handler,
	}, nil
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

// Javascript merely renders our one JS file
func Javascript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	libraries, err := webFile("react_and_axios.js")
	if err != nil {
		fail(err, w)
		return
	}
	application, err := webFile("app.js")
	if err != nil {
		fail(err, w)
		return
	}
	content := []byte(strings.Join(
		[]string{
			string(libraries),
			string(application),
		},
		";\n",
	))
	w.WriteHeader(200)
	w.Write(content)
}

// ListArchives retrieves all HAR files
func ListArchives(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

// CreateArchive stores a new HAR
func CreateArchive(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fail(err, w)
		return
	}

	archive, err := persistence.Archive{}.FromJSON(body)
	//archive, err := persistence.MakeArchive(harParam.Name, harParam.Description, har)
	if err != nil {
		fail(err, w)
		return
	}
	if err = archive.Create(db); err != nil {
		fail(err, w)
		return
	}
	// Reload it from the database to ensure the frontend always gets datastore-casted values.
	archive, err = persistence.Archive{}.Get(db, archive.ID)
	if err != nil {
		fail(err, w)
		return
	}

	w.Write(archive.AsJSON())
}

// CreateTransform stores a new transform for a specific archive
func CreateTransform(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fail(err, w)
		return
	}

	transform, err := persistence.Transform{}.FromJSON(body)
	if err != nil {
		fail(err, w)
		return
	}
	if err = transform.Create(db); err != nil {
		fail(err, w)
		return
	}
	// Reload it from the database to ensure the frontend always gets datastore-casted values.
	transform, err = persistence.Transform{}.Get(db, transform.ID)
	if err != nil {
		fail(err, w)
		return
	}

	w.Write(transform.AsJSON())
}

// UpdateArchive modifies an existing archive
func UpdateArchive(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if params["id"] == "" {
		w.WriteHeader(404)
		w.Write([]byte(`{"success": "nope"}`))
	}
	// parse this as base10 into an int64
	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		fail(err, w)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fail(err, w)
		return
	}
	archive, err := persistence.Archive{}.FromJSON(body)

	archive.ID = id

	rowsChanged, err := db.Update(&archive)
	if err != nil {
		fail(err, w)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(`{"success": "sure", "updated": %d"}`, rowsChanged)))
}

// DeleteArchive removes an archive from the db
func DeleteArchive(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if params["id"] == "" {
		w.WriteHeader(404)
		w.Write([]byte(`{"success": "nope"}`))
	}
	// parse this as base10 into an int64
	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		fail(err, w)
		return
	}

	rowsChanged, err := db.Delete(&persistence.Archive{ID: id})
	if err != nil {
		fail(err, w)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(`{"success": "sure", "deleted": %d"}`, rowsChanged)))
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

	contentJSON, err := parser.HarToJSON(har)
	if err != nil {
		fail(err, w)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(contentJSON))
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
