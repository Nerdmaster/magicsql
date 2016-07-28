package magicsql_test

import (
	"fmt"
	"github.com/Nerdmaster/magicsql"
	_ "github.com/mattn/go-sqlite3"
)

func ExampleSelect_EachRow() {
	var db, err = magicsql.Open("sqlite3", "./test.db")
	if err != nil {
		panic(err)
	}
	var op = db.Operation()

	op.Exec("DROP TABLE IF EXISTS people; CREATE TABLE people (name TEXT, age INT)")
	op.Exec("INSERT INTO people VALUES ('Joe', 100), ('Jill', 101), ('Doug', 102), ('Deb', 103)")

	var person struct {
		Name string
		Age  int
	}
	op.Select("people", &person).EachRow(func(r magicsql.Scannable) {
		r.Scan(&person.Name, &person.Age)
		fmt.Printf("%s is %d years old\n", person.Name, person.Age)
		fmt.Printf("That's %d in dog years!\n", person.Age*7)
	})

	// Output:
	// Joe is 100 years old
	// That's 700 in dog years!
	// Jill is 101 years old
	// That's 707 in dog years!
	// Doug is 102 years old
	// That's 714 in dog years!
	// Deb is 103 years old
	// That's 721 in dog years!
}

func ExampleSelect_EachObject() {
	var db, err = magicsql.Open("sqlite3", "./test.db")
	if err != nil {
		panic(err)
	}
	var op = db.Operation()

	op.Exec("DROP TABLE IF EXISTS people; CREATE TABLE people (name TEXT, age INT)")
	op.Exec("INSERT INTO people VALUES ('Joe', 100), ('Jill', 101), ('Doug', 102), ('Deb', 103)")

	var person struct {
		Name string
		Age  int
	}

	// The "&person" duplication is a necessary evil - we have to tell the Select
	// shortcut what structure to magick up, and then we have to tell EachObject
	// where to put its data
	op.Select("people", &person).EachObject(&person, func() {
		fmt.Printf("%s is %d dog years old\n", person.Name, person.Age*7)
		fmt.Printf("That's %d in people years!\n", person.Age)
	})

	// This makes more sense when using an OperationTable rather than a
	// one-off select.  For instance:
	var t = op.Table("people", &person)
	person.Name = "Pat"
	person.Age = 1
	t.Insert(&person)

	t.Select().Where("Age < ?", 100).EachObject(&person, func() {
		fmt.Printf("%s is only %d dog years old\n", person.Name, person.Age*7)
	})

	// Output:
	// Joe is 700 dog years old
	// That's 100 in people years!
	// Jill is 707 dog years old
	// That's 101 in people years!
	// Doug is 714 dog years old
	// That's 102 in people years!
	// Deb is 721 dog years old
	// That's 103 in people years!
	// Pat is only 7 dog years old
}
