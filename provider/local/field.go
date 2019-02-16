package local

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

// fieldImpl single field impl
type fieldImpl struct {
	fieldIndex int
	fieldName  string

	fieldType  typeImpl
	fieldTag   tagImpl
	fieldValue valueImpl
}

func (s *fieldImpl) GetIndex() int {
	return s.fieldIndex
}

// GetName GetName
func (s *fieldImpl) GetName() string {
	return s.fieldName
}

// GetType GetType
func (s *fieldImpl) GetType() model.FieldType {
	return &s.fieldType
}

// GetTag GetTag
func (s *fieldImpl) GetTag() model.FieldTag {
	return &s.fieldTag
}

// GetValue GetValue
func (s *fieldImpl) GetValue() model.FieldValue {
	return &s.fieldValue
}

func (s *fieldImpl) IsPrimary() bool {
	return s.fieldTag.IsPrimaryKey()
}

// Verify Verify
func (s *fieldImpl) Verify() error {
	if s.fieldTag.GetName() == "" {
		return fmt.Errorf("no define field tag")
	}

	val, err := s.fieldType.GetValue()
	if err != nil {
		return err
	}

	if s.fieldTag.IsAutoIncrement() {
		switch val {
		case util.TypeBooleanField, util.TypeStringField, util.TypeDateTimeField, util.TypeFloatField, util.TypeDoubleField, util.TypeStructField, util.TypeSliceField:
			return fmt.Errorf("illegal auto_increment field type, type:%s", s.fieldType)
		default:
		}
	}

	if s.fieldTag.IsPrimaryKey() {
		switch val {
		case util.TypeStructField, util.TypeSliceField:
			return fmt.Errorf("illegal primary key field type, type:%s", s.fieldType)
		default:
		}
	}

	return nil
}

func (s *fieldImpl) Copy() *fieldImpl {
	return &fieldImpl{
		fieldIndex: s.fieldIndex,
		fieldName:  s.fieldName,
		fieldType:  s.fieldType,
		fieldTag:   s.fieldTag,
		fieldValue: s.fieldValue,
	}
}

// Dump Dump
func (s *fieldImpl) Dump() string {
	str := fmt.Sprintf("index:[%d],name:[%s],type:[%s],tag:[%s]", s.fieldIndex, s.fieldName, s.fieldType, s.fieldTag)

	valStr, _ := s.fieldValue.Str()
	str = fmt.Sprintf("%s,value:[%s]", str, valStr)

	return str
}

func getFieldInfo(idx int, fieldType reflect.StructField) (ret *fieldImpl, err error) {
	typeImpl, err := newFieldType(fieldType.Type)
	if err != nil {
		return
	}

	tagImpl, err := newFieldTag(fieldType.Tag.Get("orm"))
	if err != nil {
		return
	}

	info := &fieldImpl{}
	info.fieldIndex = idx
	info.fieldName = fieldType.Name
	info.fieldType = *typeImpl
	info.fieldTag = *tagImpl

	err = info.Verify()
	if err != nil {
		return
	}

	ret = info
	return
}
