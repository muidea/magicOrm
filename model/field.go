package model

// Field Field
type Field interface {
	//@GetIndex Index
	GetIndex() int
	//@GetName Name
	GetName() string
	//@GetType Type
	GetType() Type
	//@GetTag Tag
	GetTag() Tag
	//@GetValue Value
	GetValue() Value
	//@SetValue 更新值
	SetValue(val Value) error
	//@IsPrimary 是否主键
	IsPrimary() bool
}

func CompareField(l, r Field) bool {
	return l.GetIndex() == r.GetIndex() &&
		l.GetName() == r.GetName() &&
		l.IsPrimary() == r.IsPrimary() &&
		CompareType(l.GetType(), r.GetType()) &&
		CompareTag(l.GetTag(), r.GetTag()) &&
		CompareValue(l.GetValue(), r.GetValue())
}

// Fields field info collection
type Fields []Field

// Append Append
func (s *Fields) Append(fieldInfo Field) bool {
	exist := false
	newField := fieldInfo.GetTag()
	for _, val := range *s {
		curField := val.GetTag()
		if curField.GetName() == newField.GetName() {
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
