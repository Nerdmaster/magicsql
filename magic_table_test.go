package magicsql

import (
	"strings"
	"testing"

	"./assert"
)

type Foo struct {
	ONE           string
	TwO           int  `sql:"two,primary"`
	Three         bool `sql:"tree"`
	Four          int
	FourPointFive int
	Five          int `sql:"-"`
	six           string
	Seven         string `sql:",readonly"`
}

func TestQueryFields(t *testing.T) {
	var table = Table("foos", &Foo{})
	assert.Equal("one,two,tree,four,four_point_five,seven", strings.Join(table.FieldNames(), ","), "Full field list", t)
	assert.Equal(6, len(table.sqlFields), "THERE ARE FOUR LIGHTS!  Er, six... and fields, not lights....", t)
}

func TestScanStruct(t *testing.T) {
	var table = Table("foos", &Foo{})
	var foo = &Foo{ONE: "blargh"}
	var ptr = table.ScanStruct(foo)[0].(*NullableField).Value.(*string)
	assert.Equal(foo.ONE, *ptr, "scanStruct properly pokes into the underlying data", t)
	*ptr = "foo"
	assert.Equal("foo", foo.ONE, "yes, this really is a proper pointer", t)
}

func TestInsertSQL(t *testing.T) {
	var table = Table("foos", &Foo{})
	var expected = "INSERT INTO foos (one,tree,four,four_point_five) VALUES (?,?,?,?)"
	assert.Equal(expected, table.InsertSQL(), "Insert SQL", t)
}

func TestInsertArgs(t *testing.T) {
	var table = Table("foos", &Foo{})
	var foo = &Foo{ONE: "blargh"}
	var save = table.InsertArgs(foo)
	assert.Equal(foo.ONE, *save[0].(*string), "Arg 1 is Foo.ONE", t)
	assert.Equal(foo.Three, *save[1].(*bool), "Arg 2 is Foo.Three since Foo.TwO is the primary key", t)
	assert.Equal(foo.Four, *save[2].(*int), "Arg 3 is Foo.Four", t)
	assert.Equal(foo.FourPointFive, *save[3].(*int), "Arg 4 is Foo.FourPointFive", t)
}

func TestUpdateSQL(t *testing.T) {
	var table = Table("foos", &Foo{})
	var expected = "UPDATE foos SET one = ?,tree = ?,four = ?,four_point_five = ? WHERE two = ?"
	assert.Equal(expected, table.UpdateSQL(), "Update SQL", t)
}

func TestUpdateArgs(t *testing.T) {
	var table = Table("foos", &Foo{})
	var foo = &Foo{ONE: "blargh"}
	var save = table.UpdateArgs(foo)
	assert.Equal(foo.ONE, *save[0].(*string), "Arg 1 is Foo.ONE", t)
	assert.Equal(foo.Three, *save[1].(*bool), "Arg 2 is Foo.Three since Foo.TwO is the primary key", t)
	assert.Equal(foo.Four, *save[2].(*int), "Arg 3 is Foo.Four", t)
	assert.Equal(foo.FourPointFive, *save[3].(*int), "Arg 4 is Foo.FourPointFive", t)
	assert.Equal(foo.TwO, *save[4].(*int), "Arg 5 is Foo.TwO (for the where clause at the end)", t)
}
