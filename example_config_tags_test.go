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

// This example showcases some of the ways SQL can be magically generated even
// without having a tagged structure
func Example_configTags() {
	// Set up a simple sqlite database
	var db, err = magicsql.Open("sqlite3", "./test.db")
	if err != nil {
		panic(err)
	}

	// Tie the "foos" table to the UntaggedFoo type
	var op = db.Operation()
	var t = op.Table("foos", &UntaggedFoo{})
	t.Reconfigure(magicsql.ConfigTags{
		"ID":    ",primary",
		"Three": "tree",
		"Five":  "-",
	})

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
	t.Save(&UntaggedFoo{ONE: "one", TwO: 2, Three: true, Four: 4})
	t.Save(&UntaggedFoo{ONE: "thing", TwO: 5, Three: false, Four: 7})
	t.Save(&UntaggedFoo{ONE: "blargh", TwO: 1, Three: true, Four: 5})
	t.Save(&UntaggedFoo{ONE: "sploop", TwO: 2, Three: true, Four: 4})
	op.EndTransaction()
	if op.Err() != nil {
		panic(op.Err())
	}

	var fooList []*UntaggedFoo
	t.Select().Where("tw_o > 1").Limit(2).Offset(1).AllObjects(&fooList)

	for _, f := range fooList {
		fmt.Printf("UntaggedFoo {%d,%s,%d,%#v,%d}\n", f.ID, f.ONE, f.TwO, f.Three, f.Four)
	}
	// Output:
	// UntaggedFoo {2,thing,5,false,7}
	// UntaggedFoo {4,sploop,2,true,4}
}
