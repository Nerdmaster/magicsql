package magicsql

import (
	"fmt"
	"reflect"
	"strings"
)

// Select defines the table, where clause, and potentially other elements of an
// in-progress SELECT statement, allowing, e.g.:
//
//     operation.From("people").Where("city = ?", varCity).SelectAllInto(peopleSlice)
//     var rows = operation.From("people").Limit(10).SelectAllRows()
type Select struct {
	ot        *OperationTable
	where     string
	whereArgs []interface{}
	order     string
	limit     uint64
	offset    uint64
}

// Where sets (or overwrites) the where clause information
func (s Select) Where(w string, args ...interface{}) Select {
	s.where = w
	s.whereArgs = args
	return s
}

// Limit sets (or overwrites) the limit value
func (s Select) Limit(l uint64) Select {
	s.limit = l
	return s
}

// Offset sets (or overwrites) the offset value
func (s Select) Offset(o uint64) Select {
	s.offset = o
	return s
}

// Order sets (or overwrites) the order clause
func (s Select) Order(o string) Select {
	s.order = o
	return s
}

// SQL returns the raw query this Select represents
func (s Select) SQL() string {
	var sql = fmt.Sprintf("SELECT %s FROM %s", strings.Join(s.ot.t.FieldNames(), ","), s.ot.t.Name)
	if s.where != "" {
		sql += fmt.Sprintf(" WHERE %s", s.where)
	}
	if s.order != "" {
		sql += fmt.Sprintf(" ORDER BY %s", s.order)
	}
	if s.limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", s.limit)
	}
	if s.offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", s.offset)
	}

	return sql
}

// AllRows builds the SQL statement, executes it through the parent
// OperationTable, and returns the resulting rows
func (s Select) AllRows() *Rows {
	if s.ot.op.Err() != nil {
		return &Rows{nil, s.ot.op}
	}

	var stmt = s.ot.op.Prepare(s.SQL())
	return stmt.Query(s.whereArgs...)
}

// EachRow wraps AllRows, yielding a Scannable per row to the callback instead
// of returning a *Rows object
func (s Select) EachRow(cb func(Scannable)) {
	var r = s.AllRows()
	defer r.Close()

	for r.Next() {
		cb(r)
	}
}

// AllObjects builds the SQL statement, executes it through the parent
// OperationTable, and returns the resulting objects into ptr, which must be a
// pointer to a slice of the type tied to this Select.
func (s Select) AllObjects(ptr interface{}) {
	var rows = s.AllRows()
	defer rows.Close()

	var slice = reflect.ValueOf(ptr).Elem()
	for rows.Next() {
		var obj = reflect.New(s.ot.t.RType).Interface()
		rows.Scan(s.ot.t.ScanStruct(obj)...)
		slice.Set(reflect.Append(slice, reflect.ValueOf(obj)))
	}
}

// EachObject mimics AllObjects, but yields each item to the callback instead
// of requiring a slice in which to put all of them at once
func (s Select) EachObject(dest interface{}, cb func()) {
	var r = s.AllRows()
	defer r.Close()

	for r.Next() {
		r.Scan(s.ot.t.ScanStruct(dest)...)
		cb()
	}
}
