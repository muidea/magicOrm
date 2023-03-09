package mysql

import (
	"fmt"
)

// EqualOpr EqualOpr
func EqualOpr(name string, val interface{}) string {
	return fmt.Sprintf("`%s` = %v", name, val)
}

// NotEqualOpr NotEqualOpr
func NotEqualOpr(name string, val interface{}) string {
	return fmt.Sprintf("`%s` != %v", name, val)
}

// BelowOpr BelowOpr
func BelowOpr(name string, val interface{}) string {
	return fmt.Sprintf("`%s` < %v", name, val)
}

// AboveOpr AboveOpr
func AboveOpr(name string, val interface{}) string {
	return fmt.Sprintf("`%s` > %v", name, val)
}

// InOpr InOpr
func InOpr(name string, val interface{}) string {
	return fmt.Sprintf("`%s` in (%v)", name, val)
}

// NotInOpr NotInOpr
func NotInOpr(name string, val interface{}) string {
	return fmt.Sprintf("`%s` not in (%v)", name, val)
}

// LikeOpr LikeOpr
func LikeOpr(name string, val interface{}) string {
	valStr, valOK := val.(string)
	if valOK {
		return fmt.Sprintf("`%s` LIKE '%%%s%%'", name, valStr[1:len(valStr)-1])
	}

	return ""
}

// SortOpr sort opr
func SortOpr(name string, ascSort bool) string {
	if ascSort {
		return fmt.Sprintf("`%s` asc", name)
	}

	return fmt.Sprintf("`%s` desc", name)
}
