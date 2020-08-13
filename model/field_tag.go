package model

// Tag Tag
type Tag interface {
	// Tag名称
	GetName() string
	// 是否是主键
	IsPrimaryKey() bool
	// 是否自增
	IsAutoIncrement() bool
}

func CompareTag(l, r Tag) bool {
	return l.GetName() == r.GetName() && l.IsPrimaryKey() == r.IsPrimaryKey() && l.IsAutoIncrement() == r.IsAutoIncrement()
}
