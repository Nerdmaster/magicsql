package magicsql

import (
	"regexp"
	"strings"
)

var re1 = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
var re2 = regexp.MustCompile(`([a-z\d])([A-Z])`)

func toUnderscore(s string) string {
	s = re1.ReplaceAllString(s, "${1}_${2}")
	s = re2.ReplaceAllString(s, "${1}_${2}")
	s = strings.Replace(s, "-", "_", -1)
	return strings.ToLower(s)
}
