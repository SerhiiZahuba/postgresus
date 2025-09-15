package sqlquery

import (
	"regexp"
	"strings"
)

// forbidden  regexp
var forbidden = regexp.MustCompile(`(?i)\b(GRANT|REVOKE|COPY|REFRESH|CALL|DO)\b`)
var multipleStmt = regexp.MustCompile(`;`)

// only SELECT/EXPLAIN/VALUES/SHOW? — : SELECT or WITH
func IsSafeSelect(sql string) bool {
	s := strings.TrimSpace(sql)
	// prohibit multiple statements
	if multipleStmt.MatchString(s) {
		return false
	}
	// prohibit modifying/DDL/utility
	if forbidden.MatchString(s) {
		return false
	}
	// allow if it starts with SELECT or WITH (CTE)
	l := strings.ToUpper(strings.TrimSpace(s))
	return strings.HasPrefix(l, "SELECT ") || strings.HasPrefix(l, "WITH ")
}