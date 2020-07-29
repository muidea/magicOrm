package model

import "reflect"

// Type Type
type Type interface {
	// Type名称
	GetName() string
	// Type值
	GetValue() int
	// Type pkgPath
	GetPkgPath() string
	// 是否指针类型
	IsPtrType() bool
	// 实例化一个类型对应的数据值
	Interface() reflect.Value
	// 获取依赖类型
	Depend() Type
	// Elem 获取聚合类型(slice)对应子项的Type，非聚合类型返回nil
	Elem() Type
}
