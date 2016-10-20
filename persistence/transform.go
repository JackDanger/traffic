package persistence

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/JackDanger/traffic/transforms"
	"github.com/JackDanger/traffic/util"
)

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

// MakeTransformFor takes a transform object (of any of the
// transform.RequestTransform implementations) and turns it into a serialized
// object that can be persisted in the database.
func MakeTransformFor(archiveID int64, transform transforms.RequestTransform) (*Transform, error) {
	// e.g. 'ConstantTransform' or 'HeaderInjectionTransform'
	transformType := strings.Split(reflect.TypeOf(transform).String(), ".")[1]
	marshaled, err := json.MarshalIndent(transform, "", "  ")
	return &Transform{
		ArchiveID:     archiveID,
		MarshaledJSON: string(marshaled),
		Type:          transformType,
	}, err
}

// Model deserializes a RequestTransform instance from the database
// MarshaledJSON string. It used repetitive `case` clauses instead of
// reflection to maintain compile-time type safety.
func (t *Transform) Model() (transforms.RequestTransform, error) {
	var instance transforms.RequestTransform
	switch t.Type {
	case "BodyToHeaderTransform":
		instance = &transforms.BodyToHeaderTransform{}
		if err := json.Unmarshal([]byte(t.MarshaledJSON), &instance); err != nil {
			return nil, err
		}
	case "HeaderToHeaderTransform":
		instance = &transforms.HeaderToHeaderTransform{}
		if err := json.Unmarshal([]byte(t.MarshaledJSON), &instance); err != nil {
			return nil, err
		}
	case "HeaderInjectionTransform":
		instance = &transforms.HeaderInjectionTransform{}
		if err := json.Unmarshal([]byte(t.MarshaledJSON), &instance); err != nil {
			return nil, err
		}
	case "ConstantTransform":
		instance = &transforms.ConstantTransform{}
		if err := json.Unmarshal([]byte(t.MarshaledJSON), &instance); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown transform type: %s", t.Type)
	}
	return instance, nil
}

// Create persists a single Transform.
func (t *Transform) Create(db *DB) error {
	if t.CreatedAt != nil {
		return errors.New("Transform already appears to be persisted")
	}

	now := util.TimePtr(time.Now())
	t.CreatedAt = now
	t.UpdatedAt = now
	err := db.Insert(t)
	return err
}

// ListTransformsFor returns all of the transform records (instantiated as
// appropriate Transform objects) for a given Archive id.
func (db *DB) ListTransformsFor(archiveID int) ([]Transform, error) {
	var records []Transform
	archiveIDColumn := db.Transforms.C("archive_id")
	err := db.Select(&records, db.Transforms.Select("*").Where(archiveIDColumn.Eq(archiveID)))
	return records, err
}

// AsJSON represents the transform as a whole in JSON.
func (t *Transform) AsJSON() []byte {
	j, _ := json.MarshalIndent(t, "", "  ")
	return j
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
