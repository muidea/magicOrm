package model

type Spec interface {
	IsPrimaryKey() bool
	GetValueDeclare() ValueDeclare
}

func CompareSpec(l, r Spec) bool {
	if l == nil && r == nil {
		return true
	}
	if l != nil && r != nil {
		return l.IsPrimaryKey() == r.IsPrimaryKey() && l.GetValueDeclare() == r.GetValueDeclare()
	}

	return false
}
