package magicsql

import (
	"strings"
	"testing"

	"./assert"
)

type Foo struct {
	ONE   string
	TwO   int
	Three bool `sql:"tree"`
	Four  int
	Five  int `sql:"-"`
	six   string
}

func newFoo() interface{} {
	return &Foo{}
}

func TestQueryFields(t *testing.T) {
	var table = NewMagicTable("foos", newFoo)
	assert.Equal("one,two,tree,four", strings.Join(table.FieldNames(), ","), "Full field list", t)
	assert.Equal(4, len(table.sqlFields), "THERE ARE FOUR LIGHTS!  Er, fields....", t)
}

func TestSQLBuilder(t *testing.T) {
	var table = NewMagicTable("foos", newFoo)
	var expected = "SELECT one,two,tree,four FROM foos WHERE x = ?"
	var actual = table.BuildQuerySQL("x = ?")
	assert.Equal(expected, actual, "SQL builder for Title", t)
}

func TestScanStruct(t *testing.T) {
	var table = NewMagicTable("foos", newFoo)
	var foo = &Foo{ONE: "blargh"}
	var ptr = table.ScanStruct(foo)[0].(*NullableField).Value.(*string)
	assert.Equal(foo.ONE, *ptr, "scanStruct properly pokes into the underlying data", t)
	*ptr = "foo"
	assert.Equal("foo", foo.ONE, "yes, this really is a proper pointer", t)
}
