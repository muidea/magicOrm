package model

import (
	"reflect"
)

// Value Value
type Value interface {
	// 是否为nil
	IsNil() bool
	// 设置值
	Set(val reflect.Value) error
	// 更新值，新旧值类型不同，则返回error
	Update(val reflect.Value) error
	// 获取值
	Get() reflect.Value
}

func CompareValue(l, r Value) bool {
	if l != nil && r != nil {
		return l.IsNil() == r.IsNil()
	}

	if l == nil && r == nil {
		return true
	}

	if l == nil && r != nil {
		return r.IsNil()
	}

	if l != nil && r == nil {
		return l.IsNil()
	}

	return false
}
