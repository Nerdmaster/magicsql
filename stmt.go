package magicsql

import (
	"database/sql"
)

// Stmt is a light wrapper for sql.Stmt
type Stmt struct {
	st *sql.Stmt
	err errorable
}

// Err returns the first error encountered on any operation the parent DB
// object oversees
func (s *Stmt) Err() error {
	return s.err.Err()
}

// Exec wraps sql.Stmt.Exec(), returning a wrapped Result
func (s *Stmt) Exec(args ...interface{}) *Result {
	if s.err.Err() != nil {
		return &Result{nil, s.err}
	}

	var r, err = s.st.Exec(args...)
	s.err.SetErr(err)
	return &Result{r, s.err}
}

// Query wraps sql.Stmt.Query(), returning a wrapped Rows
func (s *Stmt) Query(args ...interface{}) *Rows {
	if s.err.Err() != nil {
		return &Rows{nil, s.err}
	}

	var r, err = s.st.Query(args...)
	s.err.SetErr(err)
	return &Rows{r, s.err}
}

// Close wraps sql.Stmt.Close()
func (s *Stmt) Close() {
	if s.err.Err() != nil {
		return
	}

	var err = s.st.Close()
	s.err.SetErr(err)
}
