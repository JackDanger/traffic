package persistence

import (
	"encoding/json"
	"time"

	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/parser"
)

// Archive represents the database format of a single named model.Har. It's
// able to serialize and deserialize the source.
type Archive struct {
	ID        int64     `json:"-"db:"id"`
	Token     string    `json:"token"db:"token"`
	Source    string    `json:"source"db:"source"`
	CreatedAt time.Time `json:"created_at"db:"created_at"`
}

// Model deserializes a Har instance from the databse source string
func (a *Archive) Model() (*model.Har, error) {
	wrapper := &parser.HarWrapper{}
	err := json.Unmarshal([]byte(a.Source), wrapper)
	return wrapper.Har, err
}

// Schema is used to generate the table initially
func (a Archive) Schema() string {
	return `
    CREATE TABLE IF NOT EXISTS archives (
      id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
      token VARCHAR(16) NOT NULL,
      source LONGTEXT NOT NULL, -- the JSON contents of the HAR
      created_at DATETIME NOT NULL
    );`
}
