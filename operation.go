package magicsql

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

// Querier defines an interface for top-level sql types that can run SQL and
// prepare statements
type Querier interface {
	Query(string, ...interface{}) (*sql.Rows, error)
	Exec(string, ...interface{}) (sql.Result, error)
	Prepare(string) (*sql.Stmt, error)
}

// Operation represents a short-lived single-purpose combination of database
// calls.  On the first failure, its internal error is set, which all
// "children" (statements, transactions, etc) will see.  All children will
// refuse to perform any functions once an error has occurred, making it safe
// to perform a chain of related database calls and only check for an error
// when it makes sense.
//
// When a transaction is started, the operation will route all database calls
// through the transaction instead of the global database handler.  At this
// time, only one transaction at a time is supported (i.e., no nesting
// transactions).
type Operation struct {
	parent *DB
	err    error
	tx     *sql.Tx
	q      Querier
	Dbg    bool
}

// NewOperation creates an operation in its default state: its parent is the
// passed-in DB instance, and it defaults to using direct database calls until
// a transaction is started.
func NewOperation(db *DB) *Operation {
	var o = &Operation{parent: db}
	o.q = o.parent.db

	return o
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

	if op.Dbg {
		log.Printf("DEBUG - Querying: %s, %#v", query, args)
	}

	var r, err = op.q.Query(query, args...)
	op.SetErr(err)
	return &Rows{r, op}
}

// Exec wraps sql's DB.Exec, returning a wrapped Result
func (op *Operation) Exec(query string, args ...interface{}) *Result {
	if op.Err() != nil {
		return &Result{nil, op}
	}

	if op.Dbg {
		log.Printf("DEBUG - Executing: %s, %#v", query, args)
	}

	var r, err = op.q.Exec(query, args...)
	op.SetErr(err)
	return &Result{r, op}
}

// Prepare wrap's sql's DB.Prepare, returning a wrapped Stmt.  The statement
// must be closed by the caller or eventually MySQL will run out of prepared
// statements.
func (op *Operation) Prepare(query string) *Stmt {
	if op.Err() != nil {
		return &Stmt{nil, op}
	}

	if op.Dbg {
		log.Printf("DEBUG - Preparing: %s", query)
	}

	var st, err = op.q.Prepare(query)
	op.SetErr(err)
	return &Stmt{st, op}
}

// Reset clears the error if any is present
func (op *Operation) Reset() {
	op.err = nil
}

// BeginTransaction wraps sql's Begin and uses a wrapped sql.Tx to dispatch
// Query, Exec, and Prepare calls.  When the transaction is complete, instead
// of manually rolling back or committing, simply call op.EndTransaction() and
// it will rollback / commit based on the error state.  If you need to force a
// rollback, set an error manually with Operation.SetErr().
//
// If a transaction is started while one is already in progress, the operation
// gets into an error state (i.e., nested transactions are not supported).
func (op *Operation) BeginTransaction() {
	if op.Err() != nil {
		return
	}

	if op.tx != nil {
		op.SetErr(errors.New("cannot nest transactions"))
		return
	}

	var tx, err = op.parent.db.Begin()
	op.tx = tx
	op.SetErr(err)
	op.q = tx
}

// Rollback tries to roll back the transaction even if there is no error
func (op *Operation) Rollback() {
	// If there was never a transaction due to errors, this could happen and we
	// don't want a panic
	if op.tx == nil {
		return
	}

	op.tx.Rollback()
	op.tx = nil
	op.q = op.parent.db
}

// EndTransaction commits the transaction if no errors occurred, or rolls back
// if there was an error
func (op *Operation) EndTransaction() {
	// If there was never a transaction due to errors, this could happen and we
	// don't want a panic
	if op.tx == nil {
		return
	}

	if op.Err() != nil {
		op.tx.Rollback()
	} else {
		op.SetErr(op.tx.Commit())
	}

	op.tx = nil
	op.q = op.parent.db
}

// Table creates an OperationTable tied to the given table name and reflecting
// on obj's type to auto-build certain SQL statements
func (op *Operation) Table(tableName string, obj interface{}) *OperationTable {
	var mt = &MagicTable{Object: obj, Name: tableName}
	mt.Configure(nil)
	return &OperationTable{op: op, t: mt}
}

// OperationTable allows tying a stored MagicTable to this specific operation,
// rather than passing table name and an empty interface to Select() and Save()
func (op *Operation) OperationTable(mt *MagicTable) *OperationTable {
	return &OperationTable{op: op, t: mt}
}

// Save wraps Operation.Table() and Table.Save().  It creates an INSERT or
// UPDATE statement for the given object based on whether its primary key is
// zero.  Stores any errors the database returns, and fails if obj isn't tagged
// with a primary key field.
func (op *Operation) Save(tableName string, obj interface{}) *Result {
	var emptyResult = &Result{nil, op}

	if op.Err() != nil {
		return emptyResult
	}

	var ot = op.Table(tableName, obj)
	if ot.t.primaryKey == nil {
		op.SetErr(fmt.Errorf("no primary key tagged for structure %s", ot.t.RType.Name()))
		return emptyResult
	}

	return ot.Save(obj)
}

// Select wraps Operation.Table() and Table.Select().  It creates a Select
// object for further refining.
func (op *Operation) Select(tableName string, obj interface{}) Select {
	return op.Table(tableName, obj).Select()
}
