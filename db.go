// Package magicsql is a wrapper around a database handle with added magic
// to ease common SQL operations, and reflection for automatic reading and
// writing of data
package magicsql

import (
	"database/sql"
	"reflect"
)

// DB is a wrapper for sql.DB which captures errors in a way that allows the
// caller to defer error handling until a convenient time.  This structure
// responds to most of the same functions as sql.DB, and returns compatible
// structures, but never returns an error directly.  DB.Err() can be called at
// any time to check for the first error captured.
//
// DB is not meant for concurrent use.  It should be used to gather together
// any SQL which is related to a single task.  This object should not be a
// long-living object.  Once the task is complete and errors are evaluated, the
// object's life should be considered over.
//
// TODO: if the DB isn't meant to be long-living, we need to be able to
// register tables globally.  Which will mean FindAll (and future functions)
// needs to be reconsidered.  It can't be on a table object, and in fact tables
// may need to just become internal-only constructs.  Instead, the magic might
// need to be on the Rows type so it can do something like r.MagicScan(object).
// Query and Exec are where it'll get painful since those need to exist on so
// many types.  Maybe we only do table-level insert/update magic on a Tx, so
// tx.MagicInsert(object) auto-handles the prepare and exec.  Same with
// tx.MagicUpdate(object, where, args...).
//
// It would obviously be better to allow concurrent access, but that clearly
// destroys the purpose of this approach, where the first error halts
// processing without requiring per-operation error checking.
//
// On the other hand, if this were to rescope to only be a transaction wrapper,
// and not expose any functionality until a transaction has started, the
// per-task approach makes a lot more sense, and the table registration magic
// can still be done globally.
//
// This needs thought.
type DB struct {
	db     *sql.DB
	tables []*MagicTable
	tmap   map[reflect.Type]*MagicTable
	err    error
}

type errorable interface {
	Err() error
	SetErr(error)
}

// Open attempts to connect to a database, wrapping the sql.Open call, storing
// database and the error result
func Open(driverName, dataSourceName string) *DB {
	var sqldb, err = sql.Open(driverName, dataSourceName)
	var db = Wrap(sqldb)
	db.err = err
	return db
}

// Wrap is used to create a new DB from an existing connection
func Wrap(db *sql.DB) *DB {
	return &DB{db: db, tmap: make(map[reflect.Type]*MagicTable)}
}

// Err returns the *first* error which occurred
func (db *DB) Err() error {
	return db.err
}

// SetErr tells DB to stop handling any more queries.  It shouldn't usually be
// called directly, but it can be if you need to tell the DB "here's a thing
// that may be an error; don't do any more work if it is".
func (db *DB) SetErr(err error) {
	if db.Err() != nil {
		return
	}

	db.err = err
}

// RegisterTable registers and returns a Table structure with some pre-computed
// reflection data for the given generator.  The generator must be a
// zero-argument function which simply returns the type to be used with mapping
// sql to data.  It must be safe to run the generator as needed.
//
// The structure returned by the generator must have tags for explicit table
// names, or else a lowercased version of the field name will be inferred.  Tag
// names must be in the form `sql:"field_name"`.  A field name of "-" tells the
// package to skip that field.  Non-exported fields are skipped.
func (db *DB) RegisterTable(tableName string, generator func() interface{}) *MagicTable {
  var t = &MagicTable{generator: generator, name: tableName, db: db, err: db}
	t.reflect()
	db.tables = append(db.tables, t)
	db.tmap[t.RType] = t
	return t
}

// Query wraps sql's Query to ease future adaptations
func (db *DB) Query(query string, args ...interface{}) *Rows {
	if db.Err() != nil {
		return &Rows{nil, db}
	}

	var r, err = db.db.Query(query, args...)
	db.SetErr(err)
	return &Rows{r, db}
}

// Exec wraps sql's DB.Exec, returning a wrapped result
func (db *DB) Exec(query string, args ...interface{}) *Result {
	if db.Err() != nil {
		return &Result{nil, db}
	}

	var r, err = db.db.Exec(query, args...)
	db.SetErr(err)
	return &Result{r, db}
}

// Prepare wrap's sql's DB.Prepare, returning a wrapped statement
func (db *DB) Prepare(query string) *Stmt {
	if db.Err() != nil {
		return &Stmt{nil, db}
	}

	var st, err = db.db.Prepare(query)
	db.SetErr(err)
	return &Stmt{st, db}
}

// Begin wraps sql's Begin and returns a wrapped Tx.  When the transaction is
// complete, instead of manually rolling back or committing, simply call
// tx.Done() and it will rollback / commit based on the error state.  If you
// need force rollback, set the owning DB object's error via SetErr().
func (db *DB) Begin() *Tx {
	if db.Err() != nil {
		return &Tx{nil, db}
	}

	var tx, err = db.db.Begin()
	db.SetErr(err)
	return &Tx{tx, db}
}
