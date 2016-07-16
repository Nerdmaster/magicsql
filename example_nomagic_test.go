package magicsql_test

import (
	"fmt"

	"github.com/Nerdmaster/magicsql"
	_ "github.com/mattn/go-sqlite3"
)

func getdb() *magicsql.DB {
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

	return db
}

// This example showcases how you can use magicsql to simplify database access
// without relying on magically generated SQL or structure registration
func Example_withoutMagic() {
	// Start an operation
	var db = getdb()
	var op = db.Operation()

	var count = -1
	var rows = op.Query("SELECT count(*) FROM foos")
	for rows.Next() {
		rows.Scan(&count)
	}
	fmt.Printf("Row count at start: %d\n", count)

	// Create a transaction because hey, why not?
	op.BeginTransaction()
	var stmt = op.Prepare("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)")
	stmt.Exec("one", 2, true, 4)
	stmt.Exec("thing", 5, false, 7)
	stmt.Exec("blargh", 1, true, 5)
	stmt.Exec("sploop", 2, true, 4)

	// Instead of calling commit/rollback, we let the transaction figure it out
	// based on its error state
	op.EndTransaction()

	// Create a transaction and force it to fail
	op.BeginTransaction()
	stmt = op.Prepare("INSERT INTO foos (one,two,tree,four) VALUES (?, ?, ?, ?)")
	stmt.Exec("one+", 2, true, 4)
	stmt.Exec("thing+", 5, false, 7)

	rows = op.Query("SELECT COUNT(*) FROM foos")
	count = -1
	for rows.Next() {
		rows.Scan(&count)
	}
	fmt.Println("Count in transaction:", count)

	op.SetErr(fmt.Errorf("forcing rollback"))
	op.EndTransaction()

	// Reset the error state so we can continue working
	op.Reset()

	rows = op.Query("SELECT COUNT(*) FROM foos")
	count = -1
	for rows.Next() {
		rows.Scan(&count)
	}
	fmt.Println("Count after forced rollback:", count)

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
	// Count in transaction: 6
	// Count after forced rollback: 4
	// one: one, two: 2, tree: true, four: 4
	// one: thing, two: 5, tree: false, four: 7
	// one: sploop, two: 2, tree: true, four: 4
}
