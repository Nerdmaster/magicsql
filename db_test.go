// NOTE: This file relies on Foo and newFoo from the magic table test

package magicsql

import (
	"fmt"
	"testing"

	"./assert"

	_ "github.com/mattn/go-sqlite3"
)

func getdb() *DB {
	var db, err = Open("sqlite3", "./test.db")
	var sqlStmt = `
		drop table if exists foos;
		create table foos (
			one text,
			two INTEGER PRIMARY KEY AUTOINCREMENT,
			tree bool,
			four int
		);
	`
	_, err = db.DataSource().Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	return db
}

func TestSelectAllInto(t *testing.T) {
	var db = getdb()
	var source = db.DataSource()

	source.Exec("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)", "one", 2, true, 4)
	source.Exec("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)", "thing", 5, false, 7)
	source.Exec("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)", "blargh", 1, true, 5)
	source.Exec("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)", "sploop", 4, true, 4)

	db.RegisterTable("foos", newFoo)
	var op = db.Operation()
	var fooList []*Foo
	op.From("foos").Where("one = ?", "thing").SelectAllInto(&fooList)
	if op.Err() != nil {
		t.Log(op.Err())
		t.FailNow()
	}
	assert.Equal(1, len(fooList), "Retrieved one Foo", t)
	var foo = fooList[0]
	assert.Equal("thing", foo.ONE, "Retrieved valid Foo object", t)
	assert.Equal(5, foo.TwO, "Retrieved valid Foo object", t)
	assert.Equal(false, foo.Three, "Retrieved valid Foo object", t)
	assert.Equal(7, foo.Four, "Retrieved valid Foo object", t)

	fooList = nil
	op.From("foos").Where("four = ?", 4).SelectAllInto(&fooList)
	assert.Equal(2, len(fooList), "Retrieved two Foos", t)
	var foo0 = fooList[0]
	var foo1 = fooList[1]

	assert.Equal("one", foo0.ONE, "foo0.ONE", t)
	assert.Equal("sploop", foo1.ONE, "foo1.ONE", t)
}

func TestSelectAllIntoUnregistered(t *testing.T) {
	var db = getdb()
	var op = db.Operation()
	var fooList []*Foo
	op.From("foos").Where("one = ?", "thing").SelectAllInto(&fooList)
	assert.True(op.Err() != nil, "Operation should have an error", t)
	assert.Equal("table foos not registered", op.Err().Error(), "Error message is correct", t)
	assert.Equal(0, len(fooList), "fooList is empty", t)
}

func TestSaveUnregistered(t *testing.T) {
	var db = getdb()
	var source = db.DataSource()

	source.Exec("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)", "one", 2, true, 4)

	var newFoo = &Foo{ONE: "new foo"}

	var op = db.Operation()
	var result = op.Save(newFoo)
	assert.True(op.Err() != nil, "Operation should have an error", t)
	assert.Equal("table for type Foo not registered", op.Err().Error(), "Error message is correct", t)
	assert.Equal(int64(0), result.RowsAffected(), "No rows affected", t)
}

type NoPKFoo struct {
	One int
	Two string
}

func TestSaveNoPK(t *testing.T) {
	var db = getdb()
	var source = db.DataSource()

	source.Exec("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)", "one", 2, true, 4)

	db.RegisterTable("foos", func() interface{} { return &NoPKFoo{} })
	var newFoo = &NoPKFoo{One: 1, Two: "two"}

	var op = db.Operation()
	var result = op.Save(newFoo)
	assert.True(op.Err() != nil, "Operation should have an error", t)
	assert.Equal("table for type NoPKFoo has no primary key", op.Err().Error(), "Error message is correct", t)
	assert.Equal(int64(0), result.RowsAffected(), "No rows affected", t)
}

func TestSaveExisting(t *testing.T) {
	var db = getdb()
	var source = db.DataSource()
	var fooList []*Foo
	db.RegisterTable("foos", newFoo)
	var op = db.Operation()

	source.Exec("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)", "one", 2, true, 4)
	op.From("foos").SelectAllInto(&fooList)
	assert.Equal(1, len(fooList), "We have one Foo", t)
	assert.Equal("one", fooList[0].ONE, "Its `one` field is 'one'", t)
	assert.Equal(4, fooList[0].Four, "Its `four` field is 4", t)

	var existingFoo = &Foo{ONE: "updated foo", TwO: 2, Three: false}

	var result = op.Save(existingFoo)
	assert.True(op.Err() == nil, fmt.Sprintf("Operation error (%s) is nil", op.Err()), t)
	assert.Equal(int64(1), result.RowsAffected(), "1 row was updated", t)

	fooList = nil
	op.From("foos").SelectAllInto(&fooList)
	assert.Equal(1, len(fooList), "We still have one Foo", t)
	assert.Equal("updated foo", fooList[0].ONE, "Its `one` field is now 'updated foo'", t)
	assert.Equal(0, fooList[0].Four, "Its `four` field was changed to 0 by default", t)

	assert.True(op.Err() == nil, fmt.Sprintf("Operation error (%s) is still nil", op.Err()), t)
}

func TestSaveNew(t *testing.T) {
	var db = getdb()

	db.RegisterTable("foos", newFoo)
	var newFoo = &Foo{ONE: "new foo"}

	var op = db.Operation()

	var result = op.Save(newFoo)
	assert.True(op.Err() == nil, fmt.Sprintf("Operation error (%s) is nil", op.Err()), t)
	assert.Equal(int64(1), result.RowsAffected(), "1 row was inserted", t)

	var fooList []*Foo
	op.From("foos").SelectAllInto(&fooList)
	assert.Equal(1, len(fooList), "We now have one Foo", t)
	assert.Equal("new foo", fooList[0].ONE, "fooList[0].ONE", t)
	assert.Equal(1, newFoo.TwO, "newFoo.TwO was auto-populated", t)
	assert.Equal(1, fooList[0].TwO, "fooList[0].TwO was auto-populated", t)
}
