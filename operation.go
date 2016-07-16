package magicsql

import (
	"database/sql"
	"fmt"
	"reflect"
)

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

// Save creates an INSERT or UPDATE statement for the given object based on
// whether its primary key is zero.  Stores any errors the database returns,
// and fails if obj is of an unregistered type or has no primary key defined.
func (op *Operation) Save(obj interface{}) *Result {
	var emptyResult = &Result{nil, op}

	if op.Err() != nil {
		return emptyResult
	}

	var tt = reflect.TypeOf(obj).Elem()
	var t = op.parent.findTableByType(tt)
	if t == nil {
		op.SetErr(fmt.Errorf("table for type %s not registered", tt.Name()))
		return emptyResult
	}

	if t.primaryKey == nil {
		op.SetErr(fmt.Errorf("table for type %s has no primary key", tt.Name()))
		return emptyResult
	}

	// Check for object's primary key field being zero
	var rVal = reflect.ValueOf(obj).Elem()
	var pkValField = rVal.FieldByName(t.primaryKey.Field.Name)
	// TODO: check type when reading tags so we can handle non-int PKs earlier
	if pkValField.Int() == 0 {
		var res = op.Exec(t.InsertSQL(), t.InsertArgs(obj)...)
		pkValField.SetInt(res.LastInsertId())
		return res
	}

	return op.Exec(t.UpdateSQL(), t.UpdateArgs(obj)...)
}
