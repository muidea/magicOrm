package model

type Spec interface {
	IsPrimaryKey() bool
	IsAutoIncrement() bool
}

func CompareSpec(l, r Spec) bool {
	return l.IsPrimaryKey() == r.IsPrimaryKey() && l.IsAutoIncrement() == r.IsAutoIncrement()
}
