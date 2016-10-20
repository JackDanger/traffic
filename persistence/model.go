package persistence

import (
	"encoding/json"
	"time"

	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/parser"
	"github.com/JackDanger/traffic/transforms"
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

// Transform is a database representation of one of the RequestTransform
// instances for a specific Archive record.
// The transform object is serialized to JSON and stored in the
// `marshaled_json` column and the `type` column tells us which transform
// instance to instantiate when retrieving a record.
type Transform struct {
	ID            int64      `json:"id"db:"id"`
	ArchiveID     int64      `json:"archive_id"db:"archive_id"`
	Type          string     `json:"type"db:"type"`
	MarshaledJSON string     `json:"marshaled_json"db:"marshaled_json"`
	CreatedAt     *time.Time `json:"created_at"db:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"db:"updated_at"`
}

// Model deserializes a Har instance from the database source string
func (a *Archive) Model() (*model.Har, error) {
	return parser.HarFrom(a.Source)
}

// Model deserializes a RequestTransform instance from the database
// MarshaledJSON string
func (t *Transform) Model() (transforms.RequestTransform, error) {

	var instance transforms.RequestTransform
	switch t.Type {
	case "body_to_header":
		instance = transforms.BodyToHeaderTransform{}
	case "header_to_header":
		instance = &transforms.HeaderToHeaderTransform{}
	case "header_injection":
		instance = &transforms.HeaderInjectionTransform{}
	case "constant_transform":
		instance = &transforms.ConstantTransform{}
	}
	err := json.Unmarshal([]byte(t.MarshaledJSON), instance)
	return instance, err
}

// AsJSON represents the archive as a whole in JSON. The source string
// will be escaped so that it can be used as an opaque blob of content
// (or JSON parsed at will).
func (a *Archive) AsJSON() []byte {
	j, _ := json.MarshalIndent(a, "", "  ")
	return j
}

// AsJSON represents the transform as a whole in JSON.
func (t *Transform) AsJSON() []byte {
	j, _ := json.MarshalIndent(t, "", "  ")
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

// Schema is used to generate the table initially
func (t Transform) Schema() string {
	return `
    CREATE TABLE IF NOT EXISTS transforms (
      id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
      archive_id INT NOT NULL,
      type VARCHAR(255) NOT NULL, -- the name of a Go struct
      marshaled_json TEXT NOT NULL, -- the Go struct marshaled
      created_at DATETIME NOT NULL,
      updated_at DATETIME NOT NULL
    );`
}
