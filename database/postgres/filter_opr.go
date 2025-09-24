package postgres

import (
	"fmt"
	"strings"

	"github.com/muidea/magicOrm/model"
)

type OprFunc func(string, any, *ResultStack) string

func getOprFunc(filterItem model.FilterItem) (ret OprFunc) {
	switch filterItem.OprCode() {
	case model.EqualOpr:
		return EqualOpr
	case model.NotEqualOpr:
		return NotEqualOpr
	case model.BelowOpr:
		return BelowOpr
	case model.AboveOpr:
		return AboveOpr
	case model.InOpr:
		return InOpr
	case model.NotInOpr:
		return NotInOpr
	case model.LikeOpr:
		return LikeOpr
	}

	return nil
}

// EqualOpr Equal Opr
func EqualOpr(name string, val any, resultStackPtr *ResultStack) string {
	resultStackPtr.PushArgs(val)
	return fmt.Sprintf("\"%s\" = $%d", name, len(resultStackPtr.argsVal))
}

// NotEqualOpr NotEqual Opr
func NotEqualOpr(name string, val any, resultStackPtr *ResultStack) string {
	resultStackPtr.PushArgs(val)
	return fmt.Sprintf("\"%s\" != $%d", name, len(resultStackPtr.argsVal))
}

// BelowOpr Below Opr
func BelowOpr(name string, val any, resultStackPtr *ResultStack) string {
	resultStackPtr.PushArgs(val)
	return fmt.Sprintf("\"%s\" < $%d", name, len(resultStackPtr.argsVal))
}

// AboveOpr Above Opr
func AboveOpr(name string, val any, resultStackPtr *ResultStack) string {
	resultStackPtr.PushArgs(val)
	return fmt.Sprintf("\"%s\" > $%d", name, len(resultStackPtr.argsVal))
}

// InOpr In Opr
func InOpr(name string, val any, resultStackPtr *ResultStack) string {
	sliceVal, sliceOK := val.([]any)
	if !sliceOK {
		resultStackPtr.PushArgs(val)
		return fmt.Sprintf("\"%s\" IN ($%d)", name, len(resultStackPtr.argsVal))
	}

	placeHolder := []string{}
	for _, sv := range sliceVal {
		resultStackPtr.PushArgs(sv)
		placeHolder = append(placeHolder, fmt.Sprintf("$%d", len(resultStackPtr.argsVal)))
	}

	return fmt.Sprintf("\"%s\" IN (%s)", name, strings.Join(placeHolder, ","))
}

// NotInOpr NotIn Opr
func NotInOpr(name string, val any, resultStackPtr *ResultStack) string {
	sliceVal, sliceOK := val.([]any)
	if !sliceOK {
		resultStackPtr.PushArgs(val)
		return fmt.Sprintf("\"%s\" NOT IN ($%d)", name, len(resultStackPtr.argsVal))
	}

	placeHolder := []string{}
	for _, sv := range sliceVal {
		resultStackPtr.PushArgs(sv)
		placeHolder = append(placeHolder, fmt.Sprintf("$%d", len(resultStackPtr.argsVal)))
	}

	return fmt.Sprintf("\"%s\" NOT IN (%s)", name, strings.Join(placeHolder, ","))
}

// LikeOpr Like Opr
func LikeOpr(name string, val any, resultStackPtr *ResultStack) string {
	resultStackPtr.PushArgs(fmt.Sprintf("%%%s%%", val))
	return fmt.Sprintf("\"%s\" LIKE $%d", name, len(resultStackPtr.argsVal))
}

// SortOpr sort opr
func SortOpr(name string, ascSort bool) string {
	if ascSort {
		return fmt.Sprintf("\"%s\" ASC", name)
	}

	return fmt.Sprintf("\"%s\" DESC", name)
}
