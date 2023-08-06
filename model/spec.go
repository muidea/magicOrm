package model

type Spec interface {
	IsPrimaryKey() bool
	GetValueDeclare() ValueDeclare
	IsAutoIncrement() bool
	//IsUUID() bool
	//IsSnowFlake() bool
}

func CompareSpec(l, r Spec) bool {
	if l == nil && r == nil {
		return true
	}
	if l != nil && r != nil {
		return l.IsPrimaryKey() == r.IsPrimaryKey() && l.IsAutoIncrement() == r.IsAutoIncrement()
	}

	return false
}
