package magicsql_test

import (
	"fmt"
	"github.com/Nerdmaster/magicsql"
	_ "github.com/mattn/go-sqlite3"
)

func Example_withoutMagic() {
	// Set up a simple sqlite database
	var db, err = magicsql.Open("sqlite3", "./test.db")
	if err != nil {
		panic(err)
	}

	var sqlStmt = `
		CREATE TABLE IF NOT EXISTS foos (
			id   INTEGER NOT NULL PRIMARY KEY,
			one  TEXT,
			two  INT,
			tree BOOL,
			four INT
		);
		DELETE FROM foos;
	`
	_, err = db.DataSource().Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	// Start an operation
	var op = db.Operation()

	var count = 0
	var rows = op.Query("SELECT one,two,tree,four FROM foos WHERE two > 1", true)
	for rows.Next() {
		count++
	}
	fmt.Printf("Row count at start: %d\n", count)

	// Create a transaction because hey, why not?
	var tx = op.Begin()
	var stmt = tx.Prepare("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)")
	stmt.Exec("one", 2, true, 4)
	stmt.Exec("thing", 5, false, 7)
	stmt.Exec("blargh", 1, true, 5)
	stmt.Exec("sploop", 2, true, 4)

	// Instead of calling commit/rollback, we let the transaction figure it out
	// based on its error state
	tx.Done()

	rows = op.Query("SELECT one,two,tree,four FROM foos WHERE two > 1", true)
	var one string
	var two, four int
	var tree bool
	for rows.Next() {
		rows.Scan(&one, &two, &tree, &four)
		fmt.Printf("one: %s, two: %d, tree: %#v, four: %d\n", one, two, tree, four)
	}

	// Output:
	// Row count at start: 0
	// one: one, two: 2, tree: true, four: 4
	// one: thing, two: 5, tree: false, four: 7
	// one: sploop, two: 2, tree: true, four: 4
}
