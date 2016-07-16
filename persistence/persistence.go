package persistence

import (
	"database/sql"
	"encoding/json"
	"time"

	// Unsure why all examples require this package folded into the current namespace
	_ "github.com/mattn/go-sqlite3"
	"github.com/square/squalor"

	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/parser"
	"github.com/JackDanger/traffic/util"
)

var schema = `CREATE TABLE IF NOT EXISTS archive (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                token VARCHAR(16) NULL,
                source LONGTEXT NULL, -- the JSON contents of the HAR
                created_at DATETIME NULL
							); `

// Archive represents the database format of a single named model.Har. It's
// able to serialize and deserialize the source.
type Archive struct {
	ID        int       `json:"-"`
	Token     string    `json:"token"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

// Model deserializes a Har instance from the databse source string
func (a *Archive) Model() (*model.Har, error) {
	wrapper := &parser.HarWrapper{}
	err := json.Unmarshal([]byte(a.Source), wrapper)
	return wrapper.Har, err
}

// Initialize creates and migrates the database
func Initialize() error {
	_, err := db().Query(schema)
	if err != nil {
		return err
	}
	return nil
}

var _db *squalor.DB

// A memoized, lazily-initialized database object
func db() *squalor.DB {
	if _db != nil {
		return _db
	}
	sqlite3, err := sql.Open("sqlite3", "./db.sqlite3")
	if err != nil {
		panic("Cannot open ./db.sqlite3")
	}
	_db = squalor.NewDB(sqlite3)
	_db.BindModel("archives", Archive{})
	return _db
}

// Store persists a Har to the database
func Store(har *model.Har) (*Archive, error) {
	archive := &Archive{
		Token:     util.UUID(),
		Source:    parser.HarToJSON(har),
		CreatedAt: time.Now(),
	}
	err := db().Insert(archive)
	return archive, err
}

// List returns all of the har records from the database as model instances
func List() ([]Archive, error) {
	rows, err := db().Query(`SELECT token, name, source from hars`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {

	}
	return []Archive{}, nil
}
