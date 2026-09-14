//go:build !mysql

package orm

func dialectSQLForTest(template string) string { return template }
