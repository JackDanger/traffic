package persistence

import (
	"os"
	"testing"
)

var db *DB

func TestMain(m *testing.M) {
	// Creates the database if necessary
	var err error
	db, err = NewDb()
	if err != nil {
		panic(err)
	}

	exitCode := m.Run()

	db.Truncate()

	os.Exit(exitCode)
}
