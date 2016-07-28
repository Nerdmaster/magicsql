// Package magicsql is a wrapper around a database handle with added magic
// to ease common SQL operations, and reflection for automatic reading and
// writing of data
package magicsql

import (
	"database/sql"
)

type errorable interface {
	Err() error
	SetErr(error)
}

// ConfigTags is a string-to-string map for treating untagged structures as if
// they were tagged at runtime
type ConfigTags map[string]string

// DB wraps an sql.DB, providing the Operation spawner for deferred-error
// database operations.  Like sql.DB, this DB type is meant to live more or
// less globally and be a long-living object.
type DB struct {
	db *sql.DB
}

// Open attempts to connect to a database, wrapping the sql.Open call,
// returning the new magicsql.DB and error if any.  This isn't storing the
// error for later, as there's nothing which can happen if the database can't
// be opened.
func Open(driverName, dataSourceName string) (*DB, error) {
	var sqldb, err = sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return Wrap(sqldb), nil
}

// Wrap is used to create a new DB from an existing connection
func Wrap(db *sql.DB) *DB {
	return &DB{db: db}
}

// DataSource returns the underlying sql.DB pointer so the caller can do
// lower-level work which isn't wrapped in this package
func (db *DB) DataSource() *sql.DB {
	return db.db
}

// Operation returns an Operation instance, suitable for a short-lived task.
// This is the entry point for any of the sql wrapped magic.  An Operation
// should be considered a short-lived object which is not safe for concurrent
// access since it needs to be able to halt on any error with any operation it
// performs.  Concurrent access could be extremely confusing in this context
// due to the possibility of Operation.Err() returning an error from a
// different goroutine than the one doing the checking.
func (db *DB) Operation() *Operation {
	return NewOperation(db)
}
