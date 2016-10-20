package persistence

import (
	"database/sql"
	"fmt"

	// The act of importing a database/sql driver modifies database/sql, you
	// don't need to reference it unless you need access to things like
	// mysql.Error
	"github.com/go-sql-driver/mysql"
	"github.com/square/squalor"

	"github.com/JackDanger/traffic/util"
)

// DB is our wrapper around a Squalor connection. We define a few methods but
// it delegates everything else to Squalor.
type DB struct {
	*squalor.DB

	Archives   *squalor.Model
	Transforms *squalor.Model
}

// Model is what we'll call any type that represents the individual records in
// the database.
type Model interface {
	AsJSON() []byte
	Create(*DB) error
	Schema() string
}

var _ Model = &Archive{}
var _ Model = &Transform{}

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
	for _, table := range []string{"archives", "transforms"} {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			panic(err)
		}
	}
}
