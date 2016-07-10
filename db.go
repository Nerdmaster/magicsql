// Package magicsql is a wrapper around a database handle with added magic
// to ease common SQL operations, and reflection for automatic reading and
// writing of data
package magicsql

import (
	"database/sql"
	"fmt"
	"reflect"
	"sync"
)

// DB wraps an sql.DB, providing the Operation spawner for deferred-error
// database operations.  Like sql.DB, this DB type is meant to live more or
// less globally and be a long-living object.
type DB struct {
	db      *sql.DB
	typemap map[reflect.Type]*magicTable
	namemap map[string]*magicTable
	m       sync.RWMutex
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
	return &DB{
		db:      db,
		typemap: make(map[reflect.Type]*magicTable),
		namemap: make(map[string]*magicTable),
	}
}

// DataSource returns the underlying sql.DB pointer so the caller can do
// lower-level work which isn't wrapped in this package
func (db *DB) DataSource() *sql.DB {
	return db.db
}

// RegisterTable registers a table structure using NewMagicTable, then stores
// the table for lookup in an Operation's helper functions
func (db *DB) RegisterTable(tableName string, generator func() interface{}) *magicTable {
	var t = NewMagicTable(tableName, generator)

	db.m.Lock()
	defer db.m.Unlock()
	db.typemap[t.RType] = t
	db.namemap[tableName] = t

	return t
}

// findTableByName looks up the table by its name, for use in creating SQL queries
// and putting data into structures
func (db *DB) findTableByName(tableName string) *magicTable {
	db.m.RLock()
	defer db.m.RUnlock()

	return db.namemap[tableName]
}

// Operation returns an Operation instance, suitable for a short-lived task.
// This is the entry point for any of the sql wrapped magic.  An Operation
// should be considered a short-lived object which is not safe for concurrent
// access since it needs to be able to halt on any error with any operation it
// performs.  Concurrent access could be extremely confusing in this context
// due to the possibility of Operation.Err() returning an error from a
// different goroutine than the one doing the checking.
func (db *DB) Operation() *Operation {
	return &Operation{parent: db, db: db.db}
}

// Operation represents a short-lived single-purpose combination of database
// calls.  On the first failure, its internal error is set, which all
// "children" (statements, transactions, etc) will see.  All children will
// refuse to perform any functions once an error has occurred, making it safe
// to perform a chain of related database calls and only check for an error
// when it makes sense.
type Operation struct {
	parent *DB
	db     *sql.DB
	err    error
}

type errorable interface {
	Err() error
	SetErr(error)
}

// Err returns the *first* error which occurred on any database call owned by the Operation
func (op *Operation) Err() error {
	return op.err
}

// SetErr tells the Operation to stop handling any more queries.  It shouldn't
// usually be called directly, but it can be if you need to tell the object
// "here's a thing that may be an error; don't do any more work if it is".
func (op *Operation) SetErr(err error) {
	if op.Err() != nil {
		return
	}

	op.err = err
}

// Query wraps sql's Query, returning a wrapped Rows object
func (op *Operation) Query(query string, args ...interface{}) *Rows {
	if op.Err() != nil {
		return &Rows{nil, op}
	}

	var r, err = op.db.Query(query, args...)
	op.SetErr(err)
	return &Rows{r, op}
}

// Exec wraps sql's DB.Exec, returning a wrapped Result
func (op *Operation) Exec(query string, args ...interface{}) *Result {
	if op.Err() != nil {
		return &Result{nil, op}
	}

	var r, err = op.db.Exec(query, args...)
	op.SetErr(err)
	return &Result{r, op}
}

// Prepare wrap's sql's DB.Prepare, returning a wrapped Stmt
func (op *Operation) Prepare(query string) *Stmt {
	if op.Err() != nil {
		return &Stmt{nil, op}
	}

	var st, err = op.db.Prepare(query)
	op.SetErr(err)
	return &Stmt{st, op}
}

// Begin wraps sql's Begin and returns a wrapped Tx.  When the transaction is
// complete, instead of manually rolling back or committing, simply call
// tx.Done() and it will rollback / commit based on the error state.  If you
// need to force a rollback, set an error manually with Operation.SetErr().
func (op *Operation) Begin() *Tx {
	if op.Err() != nil {
		return &Tx{nil, op}
	}

	var tx, err = op.db.Begin()
	op.SetErr(err)
	return &Tx{tx, op}
}

// From starts a scoped SQL chain for firing off a SELECT against the given
// table, allowing things like this:
//
//     operation.From("people").Where("city = ?", varCity).SelectAllInto(peopleSlice)
//     var rows = operation.From("people").Limit(10).SelectAllRows()
//
// A scope can be passed around, but right now its capabilities are extremely
// limited, and I don't plan to make this particularly impressive.
func (op *Operation) From(tableName string) *Select {
	var emptySelect = &Select{parent: op}
	if op.Err() != nil {
		return emptySelect
	}

	var t = op.parent.findTableByName(tableName)
	if t == nil {
		op.SetErr(fmt.Errorf("table %s not registered", tableName))
		return emptySelect
	}

	return newSelect(op, t)
}
