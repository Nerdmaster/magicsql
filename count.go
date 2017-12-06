package magicsql

import "fmt"

// Counter is a one-off type for mapping SELECT COUNT(...) statements
type Counter struct {
	RowCount uint64
}

// Count is a subset of a Select used explicitly for counting records
type Count struct {
	Select
	counter *Counter
}

func countSQL(s Select) string {
	var sql = fmt.Sprintf("SELECT COUNT(*) FROM %s", s.ot.t.Name)
	if s.where != "" {
		sql += fmt.Sprintf(" WHERE %s", s.where)
	}

	return sql
}

// Count returns a count object built from this select statement
func (s Select) Count() Count {
	var counter = &Counter{}
	var countOT = s.ot.op.Table(s.ot.t.Name, counter)
	var c = Count{Select: s, counter: counter}
	c.ot = countOT
	c.sqlfunc = countSQL
	return c
}

func (c Count) RowCount() uint64 {
	c.First(c.counter)
	return c.counter.RowCount
}
