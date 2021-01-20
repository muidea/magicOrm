package model

// Type Type
type Type interface {
	// @GetName 名称
	GetName() string
	// @GetValue 值
	GetValue() int
	// @GetPkgPath pkgPath
	GetPkgPath() string
	// @IsPtrType 是否指针类型
	IsPtrType() bool
	// @Interface 实例化一个类型对应的数据值
	Interface() (Value, error)
	// Elem 获取要素类型(如果非slice，则返回的是本身，如果是slice,则返回slice的elem类型)
	Elem() Type
	// @IsBasic 判断是否基础类型(不是struct，也不是slice struct)
	IsBasic() bool
}

func CompareType(l, r Type) bool {
	return l.GetName() == r.GetName() && l.GetValue() == r.GetValue() && l.GetPkgPath() == r.GetPkgPath() && l.IsPtrType() == r.IsPtrType()
}
