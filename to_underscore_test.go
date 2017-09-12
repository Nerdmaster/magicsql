package magicsql

import (
	"testing"

	"./assert"
)

func TestToUnderscore(t *testing.T) {
	assert.Equal("one", toUnderscore("ONE"), "We aren't putting underscores between a single word's letters", t)
	assert.Equal("one_two", toUnderscore("OneTwo"), "OneTwo == one_two", t)
	assert.Equal("one_two", toUnderscore("oneTwo"), "oneTwo == one_two", t)
	assert.Equal("one_two", toUnderscore("ONETwo"), "ONETwo == one_two", t)
	assert.Equal("job_id", toUnderscore("JobID"), "JobID == job_id, not job_i_d or something dumb", t)
}
