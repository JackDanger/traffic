package server

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/persistence"
	"github.com/JackDanger/traffic/util"
)

func TestMain(m *testing.M) {
	// Creates the database if necessary
	var err error
	// `db` is defined in server.go
	db, err = persistence.NewDb()
	if err != nil {
		panic(err)
	}

	exitCode := m.Run()

	db.Truncate()

	os.Exit(exitCode)
}
func TestCreateArchive(t *testing.T) {
	har := util.Fixture()
	har.Entries = []model.Entry{}
	archive, err := persistence.MakeArchive("name", "description", &har)
	if err != nil {
		t.Fatal(err)
	}
	postBody := bytes.NewBuffer(archive.AsJSON())

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/archives", postBody)
	if err != nil {
		t.Fatal(err)
	}

	CreateArchive(resp, req)

	// Find the archive from the db
	var records []persistence.Archive
	err = db.Select(&records, db.Archives.Select("*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("Expected just one record, found %d", len(records))
	}
	record := records[0]
	if record.Name != "name" {
		t.Errorf("Archive created with wrong name: %s", record.Name)
	}
	if record.Description != "description" {
		t.Errorf("Archive created with wrong description: %s", record.Description)
	}

	// Ensure the returned archive matches
	responseBody, err := ioutil.ReadAll(resp.Result().Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(responseBody) != string(record.AsJSON()) {
		t.Error(string(responseBody))
		t.Error(string(record.AsJSON()))
		t.Error("Rendered something other than the actual record in question")
	}
}
