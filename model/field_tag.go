package model

type Tag interface {
	GetName() string
	IsPrimaryKey() bool
	IsAutoIncrement() bool
}

func CompareTag(l, r Tag) bool {
	return l.GetName() == r.GetName() && l.IsPrimaryKey() == r.IsPrimaryKey() && l.IsAutoIncrement() == r.IsAutoIncrement()
}
