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

type Foo2 struct {
	ID int `sql:",primary"`
	NoInsert string `sql:",noinsert"`
	NoUpdate string `sql:",noupdate"`
}

func TestNoInsertTag(t *testing.T) {
	var table = Table("foo2s", &Foo2{})
	var foo2 = &Foo2{NoInsert: "I won't be there on creation", NoUpdate: "I will!"}

	var actual = table.InsertSQL()
	var expected = "INSERT INTO foo2s (no_update) VALUES (?)"
	assert.Equal(expected, actual, "Insert SQL for Foo2 tagged struct", t)
	var save = table.InsertArgs(foo2)
	assert.Equal(1, len(save), "Only one insert arg for Foo2", t)
	assert.Equal("I will!", *save[0].(*string), "Arg 1 is the update-only arg", t)
}

func TestNoUpdateTag(t *testing.T) {
	var table = Table("foo2s", &Foo2{})
	var foo2 = &Foo2{NoInsert: "I'll be there on update!", NoUpdate: "I won't!"}

	var actual = table.UpdateSQL()
	var expected = "UPDATE foo2s SET no_insert = ? WHERE id = ?"
	assert.Equal(expected, actual, "Update SQL for Foo2 tagged struct", t)
	var save = table.UpdateArgs(foo2)
	assert.Equal(2, len(save), "Only one update arg for Foo2 (and the id)", t)
	assert.Equal("I'll be there on update!", *save[0].(*string), "Arg 1 is the insert-only arg", t)
	assert.Equal(0, *save[1].(*int), "Arg 2 is the id", t)
}
