package mysql

import (
	"fmt"
)

// EqualOpr EqualOpr
func EqualOpr(name string, val string) string {
	return fmt.Sprintf("`%s` = %s", name, val)
}

// NotEqualOpr NotEqualOpr
func NotEqualOpr(name string, val string) string {
	return fmt.Sprintf("`%s` != %s", name, val)
}

// BelowOpr BelowOpr
func BelowOpr(name string, val string) string {
	return fmt.Sprintf("`%s` < %s", name, val)
}

// AboveOpr AboveOpr
func AboveOpr(name string, val string) string {
	return fmt.Sprintf("`%s` > %s", name, val)
}

// InOpr InOpr
func InOpr(name string, val string) string {
	return fmt.Sprintf("`%s` in (%v)", name, val)
}

// NotInOpr NotInOpr
func NotInOpr(name string, val string) string {
	return fmt.Sprintf("`%s` not in (%v)", name, val)
}

// LikeOpr LikeOpr
func LikeOpr(name string, val string) string {
	return fmt.Sprintf("`%s` LIKE '%%%s%%'", name, val[1:len(val)-1])
}
