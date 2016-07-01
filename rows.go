package magicsql

import (
	"database/sql"
)

// Rows is a light wrapper for sql.Rows
type Rows struct {
	rows *sql.Rows
	err  errorable
}

// Err returns the first error encountered on any operation the parent DB
// object oversees
func (r *Rows) Err() error {
	return r.err.Err()
}

// Next wraps sql.Rows.Next().  If an error exists, false is returned.
func (r *Rows) Next() bool {
	if r.err.Err() != nil {
		return false
	}

	return r.rows.Next()
}

// Columns wraps sql.Rows.Columns().  If an error exists, nil is returned.
func (r *Rows) Columns() []string {
	if r.err.Err() != nil {
		return nil
	}

	var cols, err = r.rows.Columns()
	r.err.SetErr(err)
	return cols
}

// Scan wraps sql.Rows.Scan()
func (r *Rows) Scan(dest ...interface{}) {
	if r.err.Err() != nil {
		return
	}

	r.err.SetErr(r.rows.Scan(dest...))
}

// Close wraps sql.Rows.Close()
func (r *Rows) Close() {
	if r.err.Err() != nil {
		return
	}

	r.err.SetErr(r.rows.Close())
}
