package persistence

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	// The act of importing a database/sql driver modifies database/sql, you
	// don't need to reference it.
	_ "github.com/go-sql-driver/mysql"
	"github.com/square/squalor"

	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/parser"
	"github.com/JackDanger/traffic/util"
)

// DB is our wrapper around a Squalor connection. We define a few methods but
// it delegates everything else to Squalor.
type DB struct {
	*squalor.DB
}

var schema = `CREATE TABLE IF NOT EXISTS archives (
                id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
                token VARCHAR(16) NULL,
                source LONGTEXT NULL, -- the JSON contents of the HAR
                created_at DATETIME NULL
							); `

// NewDb returns an instance of a single connection to the database. It's the
// handle we use for performing every database operation.
func NewDb(environment string) (*DB, error) {
	databaseName := fmt.Sprintf("traffic_%s", environment)
	// We use Sqlite3 as our datastore for now
	mysql, err := sql.Open("mysql", "@/mysql")
	if err != nil {
		return nil, err
	}

	// Create the database if necessary
	rows, err := mysql.Query(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", databaseName))
	defer rows.Close()
	if err != nil {
		fmt.Printf("error creating %s: %s\n", databaseName, err)
		return nil, err
	}
	rows.Next() // This is the line that actually persists the DDL statement, for some reason (???)

	// Create the tables if necessary
	rows, err = mysql.Query(schema)
	defer rows.Close()
	if err != nil {
		fmt.Printf("error migrating: %s\n", err)
		return nil, err
	}
	rows.Next() // This is the line that actually persists the DDL statement, for some reason (???)

	// Wrap the Sqlite3 connection in the Squalor ORM
	db := squalor.NewDB(mysql)

	// Connect specific tables to specific struct types
	if x, err := db.BindModel("archives", Archive{}); err != nil {
		fmt.Printf("x: %#v, err: %#v / %+v", x, err, err)
	}

	return &DB{db}, nil
}

// Archive represents the database format of a single named model.Har. It's
// able to serialize and deserialize the source.
type Archive struct {
	ID        int       `json:"-"db:"id"`
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

// Store persists a Har to the database
func (db *DB) Store(har *model.Har) (*Archive, error) {
	archive := &Archive{
		Token:     util.UUID(),
		Source:    parser.HarToJSON(har),
		CreatedAt: time.Now(),
	}
	err := db.Insert(archive)
	return archive, err
}

// List returns all of the har records from the database as model instances
func (db *DB) List() ([]Archive, error) {
	rows, err := db.Query(`SELECT token, name, source from hars`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {

	}
	return []Archive{}, nil
}
