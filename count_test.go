package magicsql

import (
	"testing"

	"github.com/Nerdmaster/magicsql/assert"
)

func TestSelectCount(t *testing.T) {
	var op = &Operation{}
	var s = op.Select("foos", &Foo{}).Where("x = ?", 1).Limit(10).Count()
	assert.Equal(s.SQL(), "SELECT COUNT(*) FROM foos WHERE x = ?", "SQL", t)
}
