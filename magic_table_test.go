package magicsql

import (
	"strings"
	"testing"

	"./assert"
)

type Foo struct {
	ONE   string
	TwO   int `sql:",primary"`
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

func TestSaveFieldNames(t *testing.T) {
	var table = NewMagicTable("foos", newFoo)
	assert.Equal("one,tree,four", strings.Join(table.SaveFieldNames(), ","), "Save field list", t)
	assert.Equal(3, len(table.SaveFieldNames()), "THERE ARE FOUR LIGHTS!  Er, three.  And fields, not lights.", t)
}

func TestScanStruct(t *testing.T) {
	var table = NewMagicTable("foos", newFoo)
	var foo = &Foo{ONE: "blargh"}
	var ptr = table.ScanStruct(foo)[0].(*NullableField).Value.(*string)
	assert.Equal(foo.ONE, *ptr, "scanStruct properly pokes into the underlying data", t)
	*ptr = "foo"
	assert.Equal("foo", foo.ONE, "yes, this really is a proper pointer", t)
}

func TestInsertSQL(t *testing.T) {
	var table = NewMagicTable("foos", newFoo)
	assert.Equal("INSERT INTO foos (one,tree,four) VALUES (?,?,?)", table.InsertSQL(), "Insert SQL", t)
}

func TestInsertArgs(t *testing.T) {
	var table = NewMagicTable("foos", newFoo)
	var foo = &Foo{ONE: "blargh"}
	var save = table.InsertArgs(foo)
	assert.Equal(foo.ONE, *save[0].(*string), "Arg 1 is Foo.ONE", t)
	assert.Equal(foo.Three, *save[1].(*bool), "Arg 2 is Foo.Three since Foo.TwO is the primary key", t)
	assert.Equal(foo.Four, *save[2].(*int), "Arg 3 is Foo.Four", t)
}

func TestUpdateSQL(t *testing.T) {
	var table = NewMagicTable("foos", newFoo)
	assert.Equal("UPDATE foos SET one = ?,tree = ?,four = ? WHERE two = ?", table.UpdateSQL(), "Update SQL", t)
}

func TestUpdateArgs(t *testing.T) {
	var table = NewMagicTable("foos", newFoo)
	var foo = &Foo{ONE: "blargh"}
	var save = table.UpdateArgs(foo)
	assert.Equal(foo.ONE, *save[0].(*string), "Arg 1 is Foo.ONE", t)
	assert.Equal(foo.Three, *save[1].(*bool), "Arg 2 is Foo.Three since Foo.TwO is the primary key", t)
	assert.Equal(foo.Four, *save[2].(*int), "Arg 3 is Foo.Four", t)
	assert.Equal(foo.TwO, *save[3].(*int), "Arg 4 is Foo.TwO (for the where clause at the end)", t)
}
