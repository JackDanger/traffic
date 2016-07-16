package persistence

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	// The act of importing a database/sql driver modifies database/sql, you
	// don't need to reference it.
	"github.com/go-sql-driver/mysql"
	"github.com/square/squalor"

	"github.com/JackDanger/traffic/model"
	"github.com/JackDanger/traffic/parser"
	"github.com/JackDanger/traffic/util"
)

// DB is our wrapper around a Squalor connection. We define a few methods but
// it delegates everything else to Squalor.
type DB struct {
	*squalor.DB
	Archives *squalor.Model
}

var schema = `CREATE TABLE IF NOT EXISTS archives (
                id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
                token VARCHAR(16) NOT NULL,
                source LONGTEXT NOT NULL, -- the JSON contents of the HAR
                created_at DATETIME NOT NULL
							); `

// NewDb returns an instance of a single connection to the database. It's the
// handle we use for performing every database operation.
func NewDb(environment string) (*DB, error) {
	databaseName := fmt.Sprintf("traffic_%s", environment)
	// We use Sqlite3 as our datastore for now
	conn, err := sql.Open("mysql", fmt.Sprintf("@/%s", databaseName))
	if err != nil {
		return nil, err
	}

	// Create the database if necessary
	maybeCreateDatabase(databaseName, conn)

	// Create the tables if necessary
	rows, err := conn.Query(schema)
	defer rows.Close()
	if err != nil {
		fmt.Printf("error migrating: %s\n", err)
		return nil, err
	}
	rows.Next() // This is the line that actually persists the DDL statement, for some reason (???)

	// Wrap the Sqlite3 connection in the Squalor ORM
	db := squalor.NewDB(conn)

	// Connect specific tables to specific struct types
	archives, err := db.BindModel("archives", Archive{})
	if err != nil {
		fmt.Printf("x: %#v, err: %#v / %+v", archives, err, err)
	}

	d := &DB{DB: db, Archives: archives}
	return d, nil
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
	var records []Archive
	err := db.Select(&records, db.Archives.Select("*"))
	return records, err
}

func maybeCreateDatabase(databaseName string, conn *sql.DB) {
	rows, err := conn.Query(fmt.Sprintf("SHOW TABLES FROM %s", databaseName))
	if err == nil {
		rows.Next()
		rows.Close()
		return
	}

	switch err.(type) {
	case *mysql.MySQLError:
		if err.(*mysql.MySQLError).Number == 0x419 {
			conn2, err := sql.Open("mysql", "@/mysql")
			rows, err = conn2.Query(fmt.Sprintf("CREATE DATABASE %s", databaseName))
			if err != nil {
				panic(err)
			}
			rows.Next()
			rows.Close()
		}
	default:
		panic(err)
	}
}
