package magicsql

import (
	"strings"
	"testing"

	"./assert"
)

type Foo struct {
	// ONE turns into "one" for field name, as we auto-lowercase anything not tagged
	ONE   string
	// TwO is the primary key, but not explicitly given a field name, so it'll be "two"
	TwO   int `sql:",primary"`
	// Three is explicitly set to "tree"
	Three bool `sql:"tree"`
	// Four is just lowercased to "four"
	Four  int
	// Five is explicitly skipped
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

func TestScanStruct(t *testing.T) {
	var table = NewMagicTable("foos", newFoo)
	var foo = &Foo{ONE: "blargh"}
	var ptr = table.ScanStruct(foo)[0].(*NullableField).Value.(*string)
	assert.Equal(foo.ONE, *ptr, "scanStruct properly pokes into the underlying data", t)
	*ptr = "foo"
	assert.Equal("foo", foo.ONE, "yes, this really is a proper pointer", t)
}
