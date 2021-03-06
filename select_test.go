// NOTE: This file relies on Foo from the magic table test

package magicsql

import (
	"testing"

	"github.com/Nerdmaster/magicsql/assert"
)

func TestSQL(t *testing.T) {
	var op = &Operation{}
	var s = op.Select("foos", &Foo{})
	var expectedBase = "SELECT one,two,tree,four,four_point_five,seven FROM foos"

	assert.Equal(expectedBase, s.SQL(), "SQL when there's no where/offset/limit", t)

	var s2 = s.Where("x = ?", 1)
	assert.Equal(expectedBase+" WHERE x = ?", s2.SQL(), "SQL when there's where", t)

	var s3 = s.Limit(10).Offset(100)
	assert.Equal(expectedBase+" LIMIT 10 OFFSET 100", s3.SQL(), "SQL when there's limit/offset", t)

	var s4 = s.Where("foo = 'bar' AND stuff").Limit(10).Offset(100)
	assert.Equal(expectedBase+" WHERE foo = 'bar' AND stuff LIMIT 10 OFFSET 100",
		s4.SQL(), "SQL when there's where/limit/offset", t)

	var s5 = s.Order("foo DESC").Limit(10).Where("x = ?", 1)
	assert.Equal(expectedBase+" WHERE x = ? ORDER BY foo DESC LIMIT 10",
		s5.SQL(), "SQL when there's an order clause mixed in", t)
}
