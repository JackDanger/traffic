package persistence

import (
	"encoding/json"
	"time"

	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/parser"
)

// Archive represents the database format of a single named model.Har.
// It's able to serialize and deserialize the source.
type Archive struct {
	ID          int64      `json:"-"db:"id"`
	Token       string     `json:"token"db:"token"`
	Name        string     `json:"name"db:"name"`
	Source      string     `json:"source"db:"source"`
	Description string     `json:"description"db:"description"`
	CreatedAt   *time.Time `json:"created_at"db:"created_at"`
}

// Model deserializes a Har instance from the database source string
func (a *Archive) Model() (*model.Har, error) {
	wrapper := &parser.HarWrapper{}
	err := json.Unmarshal([]byte(a.Source), wrapper)
	return wrapper.Har, err
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
      token VARCHAR(16) NOT NULL,
      name VARCHAR(255),
      description text NOT NULL, -- Let everybody know how to use this
      source LONGTEXT NOT NULL, -- the JSON contents of the HAR
      created_at DATETIME NOT NULL
    );`
}
