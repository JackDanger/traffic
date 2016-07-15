package persistence

import (
	"database/sql"

	// Unsure why all examples require this package folded into the current namespace
	_ "github.com/mattn/go-sqlite3"
)

var schema = `
CREATE TABLE IF NOT EXISTS archives (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token VARCHAR(16) NULL,
    source LONGTEXT NULL,
    created_at DATETIME NULL
);
`

// Initialize initializes the db with the necessary table
func Initialize() error {
	db, err := sql.Open("sqlite3", "./db.sqlite3")
	if err != nil {
		return err
	}
	_, err = db.Query(schema)
	if err != nil {
		return err
	}
	return nil
}
