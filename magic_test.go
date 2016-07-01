package magicsql

import (
	"strings"
	"testing"

	"./assert"

	_ "github.com/mattn/go-sqlite3"
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

func getdb() *DB {
	var magicsql = Open("sqlite3", "./test.db")
	var sqlStmt = `
		create table if not exists foos (
			id integer not null primary key,
			one text,
			two int,
			tree bool,
			four int
		);
		delete from foos;
	`
	magicsql.Exec(sqlStmt)
	if magicsql.Err() != nil {
		panic(magicsql.Err())
	}

	return magicsql
}

func TestQueryFields(t *testing.T) {
	var table = getdb().RegisterTable("foos", newFoo)
	assert.Equal("one,two,tree,four", strings.Join(table.fieldNames(), ","), "Full field list", t)
	assert.Equal(4, len(table.sqlFields), "THERE ARE FOUR LIGHTS!  Er, fields....", t)
}

func TestSQLBuilder(t *testing.T) {
	var table = getdb().RegisterTable("foos", newFoo)
	var expected = "SELECT one,two,tree,four FROM foos WHERE x = ?"
	var actual = table.buildQuerySQL("x = ?")
	assert.Equal(expected, actual, "SQL builder for Title", t)
}

func TestScanStruct(t *testing.T) {
	var table = getdb().RegisterTable("foos", newFoo)
	var foo = &Foo{ONE: "blargh"}
	var ptr = table.scanStruct(foo)[0].(*NullableField).Value.(*string)
	assert.Equal(foo.ONE, *ptr, "scanStruct properly pokes into the underlying data", t)
	*ptr = "foo"
	assert.Equal("foo", foo.ONE, "yes, this really is a proper pointer", t)
}

func TestFindAll(t *testing.T) {
	var db = getdb()

	db.Exec("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)", "one", 2, true, 4)
	db.Exec("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)", "thing", 5, false, 7)
	db.Exec("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)", "blargh", 1, true, 5)
	db.Exec("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)", "sploop", 2, true, 4)

	var table = db.RegisterTable("foos", newFoo)
	var fooList = table.FindAll("one = ?", "thing")
	if table.Err() != nil {
		t.Log(table.Err())
		t.FailNow()
	}
	assert.Equal(1, len(fooList), "Retrieved one Foo", t)
	var foo = fooList[0].(*Foo)
	assert.Equal("thing", foo.ONE, "Retrieved valid Foo object", t)
	assert.Equal(5, foo.TwO, "Retrieved valid Foo object", t)
	assert.Equal(false, foo.Three, "Retrieved valid Foo object", t)
	assert.Equal(7, foo.Four, "Retrieved valid Foo object", t)

	fooList = table.FindAll("four = ?", 4)
	assert.Equal(2, len(fooList), "Retrieved two Foos", t)
	var foo0 = fooList[0].(*Foo)
	var foo1 = fooList[1].(*Foo)

	assert.Equal("one", foo0.ONE, "foo0.ONE", t)
	assert.Equal("sploop", foo1.ONE, "foo1.ONE", t)
}
