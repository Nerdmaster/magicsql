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
