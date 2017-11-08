package magicsql

import (
	"fmt"
	"testing"
	"time"

	"./assert"
)

type FooTime struct {
	ID        int `sql:"id,primary"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func getdbTime() *DB {
	var db, err = Open("sqlite3", "./test.db")
	var sqlStmt = `
		drop table if exists foo_times;
		create table foo_times (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at datetime,
			updated_at string
		);
	`
	_, err = db.DataSource().Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	return db
}

func TestSaveTime(t *testing.T) {
	var db = getdbTime()
	var cr = time.Unix(1473000000, 0)
	var up = time.Unix(1474000000, 0)
	var ft = &FooTime{CreatedAt: cr, UpdatedAt: up}

	var op = db.Operation()

	var result = op.Save("foo_times", ft)
	assert.True(op.Err() == nil, fmt.Sprintf("Operation error (%s) is nil", op.Err()), t)
	assert.Equal(int64(1), result.RowsAffected(), "1 row was inserted", t)

	var fooTimeList []*FooTime
	op.Select("foo_times", &FooTime{}).AllObjects(&fooTimeList)
	assert.Equal(1, len(fooTimeList), "We now have one Foo", t)
	assert.Equal(1, ft.ID, "ft.ID was auto-populated", t)
	assert.Equal(1, fooTimeList[0].ID, "fooTimeList[0].ID was auto-populated", t)
	assert.Equal(cr.UTC().String(), fooTimeList[0].CreatedAt.UTC().String(),
		"fooTimeList[0].CreatedAt was stored and retrieved correctly", t)
	assert.Equal(up.UTC().String(), fooTimeList[0].UpdatedAt.UTC().String(),
		"fooTimeList[0].UpdatedAt was stored and retrieved correctly", t)

	// Are times being preserved and re-read properly?
	var ft1, ft2 []*FooTime
	op.Save("foo_times", fooTimeList[0])
	op.Select("foo_times", &FooTime{}).AllObjects(&ft1)
	op.Save("foo_times", ft1[0])
	op.Select("foo_times", &FooTime{}).AllObjects(&ft2)
	assert.Equal(time.Unix(1473000000, 0).UTC().String(), ft2[0].CreatedAt.UTC().String(),
		"After saving and reloading multiple times, created_at is still correct", t)
	assert.Equal(time.Unix(1474000000, 0).UTC().String(), ft2[0].UpdatedAt.UTC().String(),
		"After saving and reloading multiple times, updated_at is still correct", t)
}
