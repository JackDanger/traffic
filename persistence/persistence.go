package persistence

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	// The act of importing a database/sql driver modifies database/sql, you
	// don't need to reference it unless you need access to things like
	// mysql.Error
	"github.com/go-sql-driver/mysql"
	"github.com/square/squalor"

	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/parser"
	"github.com/JackDanger/traffic/transforms"
	"github.com/JackDanger/traffic/util"
)

// DB is our wrapper around a Squalor connection. We define a few methods but
// it delegates everything else to Squalor.
type DB struct {
	*squalor.DB

	Archives   *squalor.Model
	Transforms *squalor.Model
}

// NewDb returns an instance of a single connection to the database. It's the
// handle we use for performing every database operation.
func NewDb() (*DB, error) {
	return NewDbForEnv(util.EnvironmentGuess())
}

// NewDbForEnv allows us to create multiple databases for development, test,
// staging, or production from the same code base.
func NewDbForEnv(environment string) (*DB, error) {
	databaseName := fmt.Sprintf("traffic_%s", environment)

	// Connect to MySQL
	conn, err := sql.Open("mysql", fmt.Sprintf("root@/%s?parseTime=true", databaseName))
	if err != nil {
		return nil, err
	}

	// Wrap the MySQL connection in the Squalor ORM and wrap that in our own DB
	// type
	db := &DB{DB: squalor.NewDB(conn)}

	// TODO: when performance of this method becomes an issue move this to an
	// external manual step
	err = db.Migrate(databaseName)
	if err != nil {
		fmt.Println("persistence/persistence.go:49 ", err)
	}

	// Connect specific tables to specific struct types
	archives, err := db.BindModel("archives", Archive{})
	transforms, err := db.BindModel("transforms", Transform{})
	db.Archives = archives
	db.Transforms = transforms
	if err == nil {
		return db, nil
	}

	switch err.(type) {
	case *mysql.MySQLError:
		fmt.Println(err.Error())
	default:
		fmt.Println(err.Error())
	}
	return nil, err
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

// MakeTransformFor takes a transform object (of any of the
// transform.RequestTransform implementations) and turns it into a serialized
// object that can be persisted in the database.
func MakeTransformFor(archiveID int64, transform transforms.RequestTransform) (*Transform, error) {
	marshaled, err := json.MarshalIndent(transform, "", "  ")
	return &Transform{
		ArchiveID:     archiveID,
		MarshaledJSON: string(marshaled),
		Type:          strings.Split(reflect.TypeOf(transform).String(), ".")[1],
	}, err
}

// Create persists a single Archive and in a very concurrent-unsafe way
// attempts to prevent multiple insertions.
func (a *Archive) Create(db *DB) error {
	if a.CreatedAt != nil {
		return errors.New("Archive already appears to be persisted")
	}

	a.CreatedAt = util.TimePtr(time.Now())
	a.UpdatedAt = util.TimePtr(time.Now())
	err := db.Insert(a)

	return err
}

// Create persists a single Transform.
func (t *Transform) Create(db *DB) error {
	if t.CreatedAt != nil {
		return errors.New("Transform already appears to be persisted")
	}

	t.CreatedAt = util.TimePtr(time.Now())
	t.UpdatedAt = util.TimePtr(time.Now())
	err := db.Insert(t)
	return err
}

// ListArchives returns all of the har records from the database as model
// instances
func (db *DB) ListArchives() ([]Archive, error) {
	var records []Archive
	err := db.Select(&records, db.Archives.Select("*"))
	return records, err
}

// ListTransformsFor returns all of the transform records (instantiated as
// appropriate Transform objects) for a given Archive id.
func (db *DB) ListTransformsFor(archiveID int) ([]Transform, error) {
	var records []Transform
	archiveIDColumn := db.Transforms.C("archive_id")
	err := db.Select(&records, db.Transforms.Select("*").Where(archiveIDColumn.Eq(archiveID)))
	return records, err
}

// Migrate will create the database if necessary and apply necessary
// migrations.
func (db *DB) Migrate(databaseName string) error {
	// Create the tables if necessary
	err := MigrateSQL(db.DB.DB, Archive{}.Schema())
	if err == nil {
		err = MigrateSQL(db.DB.DB, Transform{}.Schema())
	}
	switch err.(type) {
	case *mysql.MySQLError:
		// "database does not exist" error
		if err.(*mysql.MySQLError).Number == 0x419 {
			// Connect to a db that we know exists and then run the CREATE DATABASE
			// query
			conn, err := sql.Open("mysql", "root@/mysql?parseTime=true")
			if err != nil {
				return err
			}
			err = MigrateSQL(conn, fmt.Sprintf("CREATE DATABASE  %s", databaseName))
			if err != nil {
				return err
			}
			// Create the tables if we just created the database
			err = MigrateSQL(db.DB.DB, Archive{}.Schema())
		}
	}
	return err
}

// MigrateSQL performs a single DDL
func MigrateSQL(conn *sql.DB, query string) error {
	rows, err := conn.Query(query)
	if err != nil {
		fmt.Printf("error migrating: %s\n", err)
		return err
	}
	rows.Next()
	rows.Close()
	return nil
}

// Truncate is a misnomer because for small (< 1 million) records "DELETE FROM
// table" is faster than "TRUNCATE table" in MySQL as TRUNCATE operates at a
// very slow O(1) and Delete is a more rapid O(n) for a very small n.
func (db *DB) Truncate() {
	db.Archives.Delete()
	db.Transforms.Delete()
}
