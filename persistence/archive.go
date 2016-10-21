package persistence

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/parser"
	"github.com/JackDanger/traffic/util"
)

// Archive represents the database format of a single named model.Har.
// It's able to serialize and deserialize the source.
type Archive struct {
	ID          int64      `json:"id"db:"id"`
	Name        string     `json:"name"db:"name"`
	Source      string     `json:"source"db:"source"`
	Description string     `json:"description"db:"description"`
	CreatedAt   *time.Time `json:"created_at"db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"db:"updated_at"`
}

// Model deserializes a Har instance from the database source string
func (a *Archive) Model() (*model.Har, error) {
	return parser.HarFrom(a.Source)
}

// MakeArchive prepares a model.Har into an Archive that can be stored.
func MakeArchive(name, description string, har *model.Har) (*Archive, error) {
	json, err := parser.HarToJSON(har)
	archive := &Archive{
		Name:        name,
		Description: description,
		Source:      json,
	}
	return archive, err
}

// FromJSON accepts the raw JSON from the frontend and Unmarshales an Archive
// instance from it. The `Source` field will still be Marshaled JSON because
// it's doubly-encoded over the wire.
func (a Archive) FromJSON(b []byte) (*Archive, error) {
	archive := &Archive{}
	if err := json.Unmarshal(b, archive); err != nil {
		return nil, err
	}
	return archive, nil
}

// Get retrieves a single record by primary key
func (a Archive) Get(db *DB, id int64) (*Archive, error) {
	archive := &Archive{}
	archive.ID = id
	err := db.Get(archive, db.Archives.C("id"))
	return archive, err
}

// Create persists a single Archive and in a very concurrent-unsafe way
// attempts to prevent multiple insertions.
func (a *Archive) Create(db *DB) error {
	if a.CreatedAt != nil {
		return errors.New("Archive already appears to be persisted")
	}
	if db == nil {
		panic("Unexpected nil database connection")
	}

	now := util.TimePtr(time.Now())
	a.CreatedAt = now
	a.UpdatedAt = now

	err := db.Insert(a)
	return err
}

// ListArchives returns all of the har records from the database as model
// instances
func (db *DB) ListArchives() ([]Archive, error) {
	var records []Archive
	err := db.Select(&records, db.Archives.Select("*"))
	return records, err
}

// AsJSON represents the archive as a whole in JSON. The source string
// will be escaped so that it can be used as an opaque blob of content
// (or JSON parsed at will).
func (a *Archive) AsJSON() []byte {
	j, _ := json.MarshalIndent(a, "", "  ")
	return j
}

// Schema is used to generate the table initially
func (a Archive) Schema() string {
	return `
    CREATE TABLE IF NOT EXISTS archives (
      id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
      name VARCHAR(255),
      description text NOT NULL, -- Let everybody know how to use this
      source LONGTEXT NOT NULL, -- the JSON contents of the HAR
      created_at DATETIME NOT NULL,
      updated_at DATETIME NOT NULL
    );`
}
