package magicsql

import (
	"database/sql"
)

// Result is a wrapper for sql.Result
type Result struct {
	r sql.Result
	err errorable
}

// Err returns the first error encountered on any operation the parent DB
// object oversees
func (r *Result) Err() error {
	return r.err.Err()
}

// LastInsertId returns the wrapped Result's LastInsertId() unless an error
// has occurred, in which case 0 is returned
func (r *Result) LastInsertId() int64 {
	if r.err.Err() != nil {
		return 0
	}

	var i, err = r.r.LastInsertId()
	r.err.SetErr(err)
	return i
}

// RowsAffected returns the wrapped Result's RowsAffected() unless an error has
// occurred, in which case 0 is returned
func (r *Result) RowsAffected() int64 {
	if r.err.Err() != nil {
		return 0
	}

	var i, err = r.r.RowsAffected()
	r.err.SetErr(err)
	return i
}
