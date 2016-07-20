package magicsql_test

import (
	"fmt"
	"github.com/Nerdmaster/magicsql"
	_ "github.com/mattn/go-sqlite3"
)

// UntaggedFoo is a mostly untagged structure, but we can still use magicsql on it via
// manual ConfigTag setup
//
// Note that the tagging on TwO will be ignored since we're providing custom
// tags - even though our custom tags won't include a mapping for "TwO".  Any
// time config is explicitly provided, struct tags are ignored in full.
type UntaggedFoo struct {
	ID    int
	ONE   string
	TwO   int `sql:"blargh"`
	Three bool
	Four  int
	Five  int
	six   string
}

// newUntaggedFoo is the generator for creating a default UntaggedFoo instance
func newUntaggedFoo() interface{} {
	return &UntaggedFoo{Five: 5, six: "six"}
}

// This example showcases some of the ways SQL can be magically generated even
// without having a tagged structure
func Example_configTags() {
	// Set up a simple sqlite database
	var db, err = magicsql.Open("sqlite3", "./test.db")
	if err != nil {
		panic(err)
	}

	// Tie the "foos" table to the UntaggedFoo type
	db.RegisterTableConfig("foos", newUntaggedFoo, magicsql.ConfigTags{
		"ID":    ",primary",
		"Three": "tree",
		"Five":  "-",
	})
	var op = db.Operation()

	// Create table schema
	op.Exec("DROP TABLE IF EXISTS foos")
	op.Exec(`
		CREATE TABLE foos (
			id INTEGER NOT NULL PRIMARY KEY,
			one TEXT,
			tw_o INT,
			tree BOOL,
			four INT
		);
	`)

	// Insert four rows
	op.BeginTransaction()
	op.Save(&UntaggedFoo{ONE: "one", TwO: 2, Three: true, Four: 4})
	op.Save(&UntaggedFoo{ONE: "thing", TwO: 5, Three: false, Four: 7})
	op.Save(&UntaggedFoo{ONE: "blargh", TwO: 1, Three: true, Four: 5})
	op.Save(&UntaggedFoo{ONE: "sploop", TwO: 2, Three: true, Four: 4})
	op.EndTransaction()
	if op.Err() != nil {
		panic(op.Err())
	}

	var fooList []*UntaggedFoo
	op.From("foos").Where("tw_o > 1").Limit(2).Offset(1).SelectAllInto(&fooList)

	for _, f := range fooList {
		fmt.Printf("UntaggedFoo {%d,%s,%d,%#v,%d,%d,%s}\n", f.ID, f.ONE, f.TwO, f.Three, f.Four, f.Five, f.six)
	}
	// Output:
	// UntaggedFoo {2,thing,5,false,7,5,six}
	// UntaggedFoo {4,sploop,2,true,4,5,six}
}
