package model

type Spec interface {
	IsPrimaryKey() bool
	IsAutoIncrement() bool
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
