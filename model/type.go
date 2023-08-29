package model

type Type interface {
	GetName() string
	GetPkgPath() string
	GetDescription() string
	GetValue() TypeDeclare
	GetPkgKey() string
	IsPtrType() bool
	Interface() Value
	// Elem 获取要素类型(如果非slice，则返回的是本身，如果是slice,则返回slice的elem类型)
	Elem() Type
	IsBasic() bool
}

func CompareType(l, r Type) bool {
	return l.GetName() == r.GetName() && l.GetValue() == r.GetValue() && l.GetPkgPath() == r.GetPkgPath() && l.IsPtrType() == r.IsPtrType()
}
