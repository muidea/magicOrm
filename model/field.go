package model

import cd "github.com/muidea/magicCommon/def"

type Field interface {
	GetName() string
	GetShowName() string
	GetDescription() string
	GetType() Type
	GetSpec() Spec
	GetValue() Value
	SetValue(val any) *cd.Error
	GetSliceValue() []Value
	AppendSliceValue(val any) *cd.Error
	Reset()
}

func CompareField(l, r Field) bool {
	return l.GetName() == r.GetName() &&
		CompareType(l.GetType(), r.GetType()) &&
		CompareSpec(l.GetSpec(), r.GetSpec()) &&
		CompareValue(l.GetValue(), r.GetValue())
}

// Fields field info collection
type Fields []Field

// Append Append
func (s *Fields) Append(vField Field) bool {
	exist := false
	newName := vField.GetName()
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

	*s = append(*s, vField)
	return true
}

// GetPrimaryField get primary key field
func (s *Fields) GetPrimaryField() Field {
	for _, val := range *s {
		if IsPrimaryField(val) {
			return val
		}
	}

	return nil
}
