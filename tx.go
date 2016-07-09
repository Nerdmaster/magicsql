package magicsql

import (
	"database/sql"
)

// Tx is a light wrapper for sql.Tx
type Tx struct {
	tx  *sql.Tx
	err errorable
}

// Err returns the first error encountered on any operation the parent DB
// object oversees
func (tx *Tx) Err() error {
	return tx.err.Err()
}

// Exec wraps sql.Tx.Exec, returning a wrapped Result
func (tx *Tx) Exec(query string, args ...interface{}) *Result {
	if tx.err.Err() != nil {
		return &Result{nil, tx.err}
	}

	var res, err = tx.tx.Exec(query, args...)
	tx.err.SetErr(err)
	return &Result{res, tx.err}
}

// Prepare wraps sql.Tx.Prepare, returning a wrapped Stmt
func (tx *Tx) Prepare(query string) *Stmt {
	if tx.err.Err() != nil {
		return &Stmt{nil, tx.err}
	}

	var st, err = tx.tx.Prepare(query)
	tx.err.SetErr(err)
	return &Stmt{st, tx.err}
}

// Query wraps sql.Tx.Query, returning a wrapped Rows
func (tx *Tx) Query(query string, args ...interface{}) *Rows {
	if tx.err.Err() != nil {
		return &Rows{nil, tx.err}
	}

	var r, err = tx.tx.Query(query, args...)
	tx.err.SetErr(err)
	return &Rows{r, tx.err}
}

// Done commits the transaction if no errors occurred, or rolls back if there
// was an error
func (tx *Tx) Done() {
	// If there was never a transaction due to errors, this could happen and we
	// don't want a panic
	if tx.tx == nil {
		return
	}

	if tx.err.Err() != nil {
		tx.tx.Rollback()
		return
	}

	tx.err.SetErr(tx.tx.Commit())
}
