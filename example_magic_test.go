package magicsql_test

import (
	"fmt"
	"github.com/Nerdmaster/magicsql"
	_ "github.com/mattn/go-sqlite3"
)

// Foo demonstrates some of the optional database magic
type Foo struct {
	// ID is the primary key, but not explicitly given a field name, so it'll be "id"
	ID int `sql:",primary"`
	// ONE turns into "one" for field name, as we auto-lowercase anything not tagged
	ONE string
	// TwO just shows that the field's case is lowercased even when it may have
	// been camelcase in the structure
	TwO int
	// Three is explicitly set to "tree"
	Three bool `sql:"tree"`
	// Four is just lowercased to "four"
	Four int
	// Five is explicitly skipped
	Five int `sql:"-"`
	// six isn't exported, so is implicitly skipped
	six string
}

// newFoo is the generator for creating a default Foo instance
func newFoo() interface{} {
	return &Foo{Five: 5, six: "six"}
}

// Example_withMagic showcases some of the ways SQL can be magically generated
// to populate registered structures
func Example_withMagic() {
	// Set up a simple sqlite database
	var db, err = magicsql.Open("sqlite3", "./test.db")
	if err != nil {
		panic(err)
	}

	var sqlStmt = `
		DROP TABLE IF EXISTS foos;
		CREATE TABLE foos (
			id   INTEGER NOT NULL PRIMARY KEY,
			one  TEXT,
			two  INT,
			tree BOOL,
			four INT
		);
		INSERT INTO foos (one,two,tree,four) VALUES ("one", 2, 1, 4);
		INSERT INTO foos (one,two,tree,four) VALUES ("thing", 5, 0, 7);
		INSERT INTO foos (one,two,tree,four) VALUES ("blargh", 1, 1, 5);
		INSERT INTO foos (one,two,tree,four) VALUES ("sploop", 2, 1, 4);
	`

	_, err = db.DataSource().Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	// Tie the "foos" table to the Foo type
	db.RegisterTable("foos", newFoo)
	var op = db.Operation()

	var fooList []*Foo
	op.From("foos").Where("two > 1").Limit(2).SelectAllInto(&fooList)

	for _, f := range fooList {
		fmt.Printf("Foo {%d,%s,%d,%#v,%d,%d,%s}\n", f.ID, f.ONE, f.TwO, f.Three, f.Four, f.Five, f.six)
	}
	// Output:
	// Foo {1,one,2,true,4,5,six}
	// Foo {2,thing,5,false,7,5,six}
}
