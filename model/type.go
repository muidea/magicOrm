package model

import cd "github.com/muidea/magicCommon/def"

type Type interface {
	GetName() string
	GetPkgPath() string
	GetPkgKey() string
	GetDescription() string
	GetValue() TypeDeclare
	IsPtrType() bool
	Interface(initVal any) (Value, *cd.Result)
	// Elem 获取要素类型(如果非slice，则返回的是本身，如果是slice,则返回slice的elem类型)
	Elem() Type
}

func CompareType(l, r Type) bool {
	return l.GetName() == r.GetName() && l.GetValue() == r.GetValue() && l.GetPkgPath() == r.GetPkgPath() && l.IsPtrType() == r.IsPtrType()
}
