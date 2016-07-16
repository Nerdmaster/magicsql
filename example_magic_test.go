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

	// Tie the "foos" table to the Foo type
	db.RegisterTable("foos", newFoo)
	var op = db.Operation()

	// Create table schema
	op.Exec("DROP TABLE IF EXISTS foos")
	op.Exec("CREATE TABLE foos (id INTEGER NOT NULL PRIMARY KEY, one TEXT, two INT, tree BOOL, four INT)")

	// Insert four rows
	op.BeginTransaction()
	op.Save(&Foo{ONE: "one", TwO: 2, Three: true, Four: 4})
	op.Save(&Foo{ONE: "thing", TwO: 5, Three: false, Four: 7})
	op.Save(&Foo{ONE: "blargh", TwO: 1, Three: true, Four: 5})
	op.Save(&Foo{ONE: "sploop", TwO: 2, Three: true, Four: 4})
	op.EndTransaction()
	if op.Err() != nil {
		panic(op.Err())
	}

	var fooList []*Foo
	op.From("foos").Where("two > 1").Limit(2).Offset(1).SelectAllInto(&fooList)

	for _, f := range fooList {
		fmt.Printf("Foo {%d,%s,%d,%#v,%d,%d,%s}\n", f.ID, f.ONE, f.TwO, f.Three, f.Four, f.Five, f.six)
	}
	// Output:
	// Foo {2,thing,5,false,7,5,six}
	// Foo {4,sploop,2,true,4,5,six}
}
