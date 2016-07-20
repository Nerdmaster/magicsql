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
	parent    *Operation
	table     *magicTable
	where     string
	whereArgs []interface{}
	order     string
	limit     uint64
	offset    uint64
}

func newSelect(p *Operation, from *magicTable) *Select {
	return &Select{parent: p, table: from}
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
	var sql = fmt.Sprintf("SELECT %s FROM %s", strings.Join(s.table.FieldNames(), ","), s.table.name)
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

// SelectAllRows builds the SQL statement, runs it against this Select's
// database Operation, and returns the resulting rows
func (s Select) SelectAllRows() *Rows {
	if s.parent.Err() != nil {
		return &Rows{nil, s.parent}
	}

	var stmt = s.parent.Prepare(s.SQL())
	return stmt.Query(s.whereArgs...)
}

// SelectAllInto builds the SQL statement, runs it against this Select's
// database Operation, and populates the slice, initializing and growing it as
// necessary.  A pointer to a registered type must be passed in:
//
//    var sl []Thing
//    op.From("table").SelectAllInto(&sl)
func (s Select) SelectAllInto(ptr interface{}) {
	var rows = s.SelectAllRows()
	var slice = reflect.ValueOf(ptr).Elem()
	for rows.Next() {
		var obj = s.table.generator()
		rows.Scan(s.table.ScanStruct(obj)...)
		slice.Set(reflect.Append(slice, reflect.ValueOf(obj)))
	}
}
