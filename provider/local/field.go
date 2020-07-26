package local

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
	"log"
	"reflect"
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
func (s *fieldImpl) GetType() model.Type {
	return &s.fieldType
}

// GetTag GetTag
func (s *fieldImpl) GetTag() model.Tag {
	return &s.fieldTag
}

// GetValue GetValue
func (s *fieldImpl) GetValue() model.Value {
	return &s.fieldValue
}

func (s *fieldImpl) IsPrimary() bool {
	return s.fieldTag.IsPrimaryKey()
}

func (s *fieldImpl) IsAssigned() (ret bool) {
	ret = false
	if s.fieldValue.IsNil() {
		return
	}

	if s.fieldType.IsPtrType() {
		ret = true
		return
	}

	currentVal := s.fieldValue.Get()
	originVal := s.fieldType.Interface()
	sameVal, sameErr := util.IsSameVal(originVal, currentVal)
	if sameErr != nil {
		log.Printf("compare value failed, err:%s", sameErr.Error())
		ret = false
		return
	}
	// 值不相等，则可以认为有赋值过
	ret = !sameVal

	return
}

func (s *fieldImpl) SetValue(val reflect.Value) (err error) {
	err = s.fieldValue.Set(val)
	return
}

func (s *fieldImpl) UpdateValue(val reflect.Value) (err error) {
	err = s.fieldValue.Update(val)

	return
}

// verify verify
func (s *fieldImpl) verify() error {
	if s.fieldTag.GetName() == "" {
		return fmt.Errorf("no define field tag")
	}

	val := s.fieldType.GetValue()
	if s.fieldTag.IsAutoIncrement() {
		switch val {
		case util.TypeBooleanField, util.TypeStringField, util.TypeDateTimeField, util.TypeFloatField, util.TypeDoubleField, util.TypeStructField, util.TypeSliceField:
			return fmt.Errorf("illegal auto_increment field type, type:%s", s.fieldType.Dump())
		default:
		}
	}

	if s.fieldTag.IsPrimaryKey() {
		switch val {
		case util.TypeStructField, util.TypeSliceField:
			return fmt.Errorf("illegal primary key field type, type:%s", s.fieldType.Dump())
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
func (s *fieldImpl) Dump(cache Cache) string {
	str, _ := getValueStr(&s.fieldType, &s.fieldValue, cache)
	str = fmt.Sprintf("index:%d,name:%s,type:[%s],tag:[%s],value:%s", s.fieldIndex, s.fieldName, s.fieldType.Dump(), s.fieldTag.Dump(), str)

	return str
}

func getFieldInfo(idx int, fieldType reflect.StructField) (ret *fieldImpl, err error) {
	typeImpl, err := newType(fieldType.Type)
	if err != nil {
		return
	}

	tagImpl, err := newTag(fieldType.Tag.Get("orm"))
	if err != nil {
		return
	}

	field := &fieldImpl{}
	field.fieldIndex = idx
	field.fieldName = fieldType.Name
	field.fieldType = *typeImpl
	field.fieldTag = *tagImpl

	err = field.verify()
	if err != nil {
		return
	}

	ret = field
	return
}
