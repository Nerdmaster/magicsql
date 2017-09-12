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
	// TwO is explicitly set so it doesn't underscorify
	TwO int `sql:"two"`
	// Three is explicitly set to "tree"
	Three bool `sql:"tree"`
	// Four is just lowercased to "four"
	Four int
	// FourPointFive gets turned into underscores
	FourPointFive int
	// Five is explicitly skipped
	Five int `sql:"-"`
	// six isn't exported, so is implicitly skipped
	six string
	// Seven is read-only, so it can be selected but not stored
	Seven string `sql:",readonly"`
	// NoInsert can't be written on insert, but can be written on update; it's
	// useful when we want the database to specify the initial value, but still
	// be able to change it later
	NoInsert int `sql:",noinsert"`
	// NoUpdate can't be written on update, but can be written on insert; it's useful
	// for values that should never change, such as a username
	NoUpdate int `sql:",noupdate"`
}

// This example showcases some of the ways SQL can be magically generated to
// populate registered structures
func Example_withMagic() {
	// Set up a simple sqlite database
	var db, err = magicsql.Open("sqlite3", "./test.db")
	if err != nil {
		panic(err)
	}

	var op = db.Operation()

	// Create table schema
	op.Exec("DROP TABLE IF EXISTS foos")
	op.Exec(`
		CREATE TABLE foos (
			id INTEGER NOT NULL PRIMARY KEY,
			one TEXT,
			two INT,
			tree BOOL,
			four INT,
			four_point_five INT,
			seven TEXT DEFAULT "blargh",
			no_insert INT,
			no_update INT
		);
	`)

	// Insert four rows
	op.BeginTransaction()
	op.Save("foos", &Foo{ONE: "one", TwO: 2, Three: true, Four: 4, FourPointFive: 9})
	op.Save("foos", &Foo{ONE: "thing", TwO: 5, Three: false, Four: 7, FourPointFive: -1})
	op.Save("foos", &Foo{ONE: "blargh", TwO: 1, Three: true, Four: 5})

	// Fields "Five" and "six" won't be preserved since there's no place to put
	// them, so we won't see their values below.  Field "Seven" is readonly and
	// so will retain its default value.  Field "Eight" won't get set on insert,
	// so will be 0 until we update.  Field "Nine" will be set on insert, but
	// can't be updated.
	op.Save("foos", &Foo{
		ONE:      "sploop",
		TwO:      2,
		Three:    true,
		Four:     4,
		Five:     29,
		six:      "twenty-nine",
		Seven:    "nope",
		NoInsert: 1010,
		NoUpdate: 1010,
	})

	op.EndTransaction()
	if op.Err() != nil {
		panic(op.Err())
	}

	var fooList []*Foo
	op.Select("foos", &Foo{}).Where("two > 1").Limit(2).Offset(1).Order("four_point_five DESC").AllObjects(&fooList)

	for _, f := range fooList {
		fmt.Printf("Foo {%d,%s,%d,%#v,%d,%d,%d,%q,%q,%d,%d}\n",
			f.ID, f.ONE, f.TwO, f.Three, f.Four, f.FourPointFive, f.Five, f.six, f.Seven, f.NoInsert, f.NoUpdate)
	}

	// Try out an update on the first result in the list, which is id 4
	var f = fooList[0]
	f.NoInsert = 99
	f.NoUpdate = 99
	op.Save("foos", f)

	op.Select("foos", &Foo{}).Where("id = ?", 4).First(f)
	fmt.Printf("Foo {%d,%s,%d,%#v,%d,%d,%d,%q,%q,%d,%d}\n",
		f.ID, f.ONE, f.TwO, f.Three, f.Four, f.FourPointFive, f.Five, f.six, f.Seven, f.NoInsert, f.NoUpdate)

	// Output:
	// Foo {4,sploop,2,true,4,0,0,"","blargh",0,1010}
	// Foo {2,thing,5,false,7,-1,0,"","blargh",0,0}
	// Foo {4,sploop,2,true,4,0,0,"","blargh",99,1010}
}
