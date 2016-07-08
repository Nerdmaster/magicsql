// NOTE: This file relies on Foo and newFoo from the magic table test

package magicsql

import (
	"testing"

	"./assert"

	_ "github.com/mattn/go-sqlite3"
)

func getdb() *DB {
	var db, err = Open("sqlite3", "./test.db")
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
	_, err = db.DataSource().Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	return db
}

func TestFindAll(t *testing.T) {
	var db = getdb()
	var source = db.DataSource()

	source.Exec("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)", "one", 2, true, 4)
	source.Exec("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)", "thing", 5, false, 7)
	source.Exec("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)", "blargh", 1, true, 5)
	source.Exec("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)", "sploop", 2, true, 4)

	db.RegisterTable("foos", newFoo)
	var op = db.Operation()
	var fooList = op.FindAll("foos", "one = ?", "thing")
	if op.Err() != nil {
		t.Log(op.Err())
		t.FailNow()
	}
	assert.Equal(1, len(fooList), "Retrieved one Foo", t)
	var foo = fooList[0].(*Foo)
	assert.Equal("thing", foo.ONE, "Retrieved valid Foo object", t)
	assert.Equal(5, foo.TwO, "Retrieved valid Foo object", t)
	assert.Equal(false, foo.Three, "Retrieved valid Foo object", t)
	assert.Equal(7, foo.Four, "Retrieved valid Foo object", t)

	fooList = op.FindAll("foos", "four = ?", 4)
	assert.Equal(2, len(fooList), "Retrieved two Foos", t)
	var foo0 = fooList[0].(*Foo)
	var foo1 = fooList[1].(*Foo)

	assert.Equal("one", foo0.ONE, "foo0.ONE", t)
	assert.Equal("sploop", foo1.ONE, "foo1.ONE", t)
}
