//go:build mysql

package orm

import (
	"regexp"
	"strings"
)

var testSQLPlaceholder = regexp.MustCompile(`\$[0-9]+`)

// Only for literal-free test SQL templates, not for transforming actual SQL.
func dialectSQLForTest(template string) string {
	return testSQLPlaceholder.ReplaceAllString(strings.ReplaceAll(template, `"`, "`"), "?")
}
