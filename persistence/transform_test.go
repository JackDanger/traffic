package persistence

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/JackDanger/traffic/parser"
	"github.com/JackDanger/traffic/transforms"
	"github.com/JackDanger/traffic/util"
)

func TestTransformCreate(t *testing.T) {
	// Create an archive to pin these transform records onto
	har, err := parser.HarFromFile(util.Root() + "fixtures/browse-two-github-users.har")
	if err != nil {
		t.Fatal(err)
	}
	archive, err := MakeArchive("some name", "any description", har)
	if err != nil {
		t.Fatal(err)
	}
	if err := archive.Create(db); err != nil {
		t.Fatal(err)
	}

	// This JSON can instantiate a transforms.RequestTransform instance
	MarshaledJSON := `{
		"type": "constant",
		"search": "UUID1",
		"replace": "abc123-xyz789"
	}`
	// Ensure it works
	transform := &transforms.ConstantTransform{}
	if err := json.Unmarshal([]byte(MarshaledJSON), transform); err != nil {
		t.Log(transform)
		t.Fatal(err)
	}
	// Create a record that represents it
	record, err := MakeTransformFor(archive.ID, transform)
	if err != nil {
		t.Fatal(err)
	}

	if err := record.Create(db); err != nil {
		t.Fatal(err)
	}
	if record == nil {
		t.Fatal("Transform record not expected to be nil")
	}
	if record.ID == 0 {
		t.Fatal("record.ID expected to be non-zero")
	}
	if record.ArchiveID != archive.ID {
		t.Errorf("Expected archive.IDe to be %d, got: %d", archive.ID, record.ArchiveID)
	}
	if record.Type != "ConstantTransform" {
		t.Errorf("Expected record.Type to be %s, got: %s", "ConstantTransform", record.Type)
	}
	// If the time since it was created is less than "-1 * time.Second" or "1 second ago"
	if record.CreatedAt.Sub(time.Now()) < -time.Second {
		t.Errorf("Unexpected record.CreatedAt: %#v, duration: %#v (%#v)", record.CreatedAt, record.CreatedAt.Sub(time.Now()), time.Second)
	}
	// The updated_at should be about the same as created_at
	if record.UpdatedAt.Sub(time.Now()) < -time.Second {
		t.Errorf("Unexpected record.UpdatedAt: %#v, duration: %#v (%#v)", record.UpdatedAt, record.UpdatedAt.Sub(time.Now()), time.Second)
	}
}

func TestTransformModel(t *testing.T) {
	// This JSON can instantiate a transforms.RequestTransform instance
	MarshaledJSON := `{
		"type": "constant",
		"search": "UUID1",
		"replace": "abc123-xyz789"
	}`
	// Ensure it works
	transform := &transforms.ConstantTransform{}
	if err := json.Unmarshal([]byte(MarshaledJSON), transform); err != nil {
		t.Fatal(err)
	}
	// Create a record that represents it
	record, err := MakeTransformFor(1, transform)
	if err != nil {
		t.Fatal(err)
	}

	// Extract the model as the type we expect
	retrievedTransform, err := record.Model()
	if err != nil {
		t.Fatal(err)
	}

	if retrievedTransform == nil {
		t.Error("Expected record.Model() to be the unmarshaled transform instance")
	}

	// Cast it to the known type to inspect the type-specific attributes
	casted := retrievedTransform.(*transforms.ConstantTransform)
	if casted.Search != "UUID1" {
		t.Fatalf("Expected casted.Search to equal %s, was %s", "UUID1", casted.Search)
	}
	if casted.Replace != "abc123-xyz789" {
		t.Fatalf("Expected casted.Search to equal %s, was %s", "abc123-xyz789", casted.Replace)
	}
}

func TestListTransformsFor(t *testing.T) {
	//db.Transforms.Select("*").Where(db.Transforms.C("archive_id").Eq(archive.ID))

	//err := db.Select(&records, db.Archives.Select("*"))
}
