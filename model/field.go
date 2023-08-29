package model

type Field interface {
	GetIndex() int
	GetName() string
	GetDescription() string
	GetType() Type
	GetSpec() Spec
	GetValue() Value
	SetValue(val Value) error
	IsPrimary() bool
}

func CompareField(l, r Field) bool {
	return l.GetIndex() == r.GetIndex() &&
		l.GetName() == r.GetName() &&
		l.IsPrimary() == r.IsPrimary() &&
		CompareType(l.GetType(), r.GetType()) &&
		CompareSpec(l.GetSpec(), r.GetSpec()) &&
		CompareValue(l.GetValue(), r.GetValue())
}

// Fields field info collection
type Fields []Field

// Append Append
func (s *Fields) Append(fieldInfo Field) bool {
	exist := false
	newName := fieldInfo.GetName()
	for _, val := range *s {
		curName := val.GetName()
		if curName == newName {
			exist = true
			break
		}
	}
	if exist {
		return false
	}

	*s = append(*s, fieldInfo)
	return true
}

// GetPrimaryField get primary key field
func (s *Fields) GetPrimaryField() Field {
	for _, val := range *s {
		if val.IsPrimary() {
			return val
		}
	}

	return nil
}
