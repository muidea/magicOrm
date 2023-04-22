package mysql

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
)

type OprFunc func(string, interface{}) string

func getOprFunc(filterItem model.FilterItem) (ret OprFunc) {
	switch filterItem.OprCode() {
	case model.Equal:
		return EqualOpr
	case model.NotEqual:
		return NotEqualOpr
	case model.Below:
		return BelowOpr
	case model.Above:
		return AboveOpr
	case model.In:
		return InOpr
	case model.NotIn:
		return NotInOpr
	case model.Like:
		return LikeOpr
	}

	return nil
}

// EqualOpr Equal Opr
func EqualOpr(name string, val interface{}) string {
	return fmt.Sprintf("`%s` = %v", name, val)
}

// NotEqualOpr NotEqual Opr
func NotEqualOpr(name string, val interface{}) string {
	return fmt.Sprintf("`%s` != %v", name, val)
}

// BelowOpr Below Opr
func BelowOpr(name string, val interface{}) string {
	return fmt.Sprintf("`%s` < %v", name, val)
}

// AboveOpr Above Opr
func AboveOpr(name string, val interface{}) string {
	return fmt.Sprintf("`%s` > %v", name, val)
}

// InOpr In Opr
func InOpr(name string, val interface{}) string {
	return fmt.Sprintf("`%s` in (%v)", name, val)
}

// NotInOpr NotIn Opr
func NotInOpr(name string, val interface{}) string {
	return fmt.Sprintf("`%s` not in (%v)", name, val)
}

// LikeOpr Like Opr
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
		return fmt.Sprintf("`%s` ASC", name)
	}

	return fmt.Sprintf("`%s` DESC", name)
}
