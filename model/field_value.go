package model

import "reflect"

// Value Value
type Value interface {
	// 是否为nil
	IsNil() bool
	// 设置值
	Set(val reflect.Value) error
	// 获取值
	Get() reflect.Value
	// 获取指针
	Addr() Value
	// 判断值是否是基础类型
	IsBasic() bool
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
