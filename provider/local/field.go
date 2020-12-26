package local

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/util"

	log "github.com/cihub/seelog"
)

// fieldImpl single field impl
type fieldImpl struct {
	fieldIndex int
	fieldName  string

	fieldType  *typeImpl
	fieldTag   *tagImpl
	fieldValue *valueImpl
}

func (s *fieldImpl) GetIndex() int {
	return s.fieldIndex
}

// GetName GetName
func (s *fieldImpl) GetName() string {
	return s.fieldName
}

// GetType GetType
func (s *fieldImpl) GetType() (ret model.Type) {
	if s.fieldType != nil {
		ret = s.fieldType
	}

	return
}

// GetTag GetTag
func (s *fieldImpl) GetTag() (ret model.Tag) {
	if s.fieldTag != nil {
		ret = s.fieldTag
	}

	return
}

// GetValue GetValue
func (s *fieldImpl) GetValue() (ret model.Value) {
	if s.fieldValue != nil {
		ret = s.fieldValue
	}

	return
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
		if util.IsStructType(s.fieldType.GetValue()) {
			ret = true
			return
		}

		if util.IsSliceType(s.fieldType.GetValue()) {
			val := reflect.Indirect(s.fieldValue.Get())
			if val.Len() > 0 {
				return true
			}

			return false
		}
	}

	currentVal := s.fieldValue.Get()
	originVal := s.fieldType.Interface()
	sameVal, sameErr := util.IsSameVal(originVal, currentVal)
	if sameErr != nil {
		ret = false
		return
	}
	// 值不相等，则可以认为有赋值过
	ret = !sameVal

	return
}

func (s *fieldImpl) SetValue(val reflect.Value) (err error) {
	err = s.fieldValue.Set(val)
	if err != nil {
		log.Errorf("set field value failed, name:%s, err:%s", s.fieldName, err.Error())
	}

	return
}

func (s *fieldImpl) UpdateValue(val reflect.Value) (err error) {
	if util.IsNil(val) {
		err = fmt.Errorf("update value is nil")
		return
	}

	if s.fieldType.IsBasic() {
		toVal := s.fieldType.Interface()
		toVal, err = helper.AssignValue(val, toVal)
		if err != nil {
			log.Errorf("assign value failed, name:%s, from type:%s, to type:%s, err:%s", s.fieldName, val.Type().String(), s.fieldType.GetName(), err.Error())
			return
		}

		err = s.fieldValue.Update(toVal)
		return
	}

	vType, vErr := newType(val.Type())
	if vErr != nil {
		err = vErr
		return
	}

	if vType.GetName() != s.fieldType.GetName() {
		err = fmt.Errorf("invalid update value, value type:%s, expect type:%s", vType.GetName(), s.fieldType.GetName())
		return
	}

	err = s.fieldValue.Update(val)
	if err != nil {
		log.Errorf("update field value failed, name:%s, err:%s", s.fieldName, err.Error())
	}

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
		case util.TypeBooleanField,
			util.TypeStringField,
			util.TypeDateTimeField,
			util.TypeFloatField,
			util.TypeDoubleField,
			util.TypeStructField,
			util.TypeSliceField:
			return fmt.Errorf("illegal auto_increment field type, type:%s", s.fieldType.dump())
		default:
		}
	}

	if s.fieldTag.IsPrimaryKey() {
		switch val {
		case util.TypeStructField, util.TypeSliceField:
			return fmt.Errorf("illegal primary key field type, type:%s", s.fieldType.dump())
		default:
		}
	}

	return nil
}

func (s *fieldImpl) copy() *fieldImpl {
	return &fieldImpl{
		fieldIndex: s.fieldIndex,
		fieldName:  s.fieldName,
		fieldType:  s.fieldType.copy(),
		fieldTag:   s.fieldTag.copy(),
		fieldValue: s.fieldValue.copy(),
	}
}

// Dump Dump
func (s *fieldImpl) dump() string {
	str := fmt.Sprintf("index:%d,name:%s,type:[%s],tag:[%s]", s.fieldIndex, s.fieldName, s.fieldType.dump(), s.fieldTag.dump())
	return str
}

func getFieldInfo(idx int, fieldType reflect.StructField, fieldValue reflect.Value) (ret *fieldImpl, err error) {
	typeImpl, typeErr := newType(fieldType.Type)
	if typeErr != nil {
		err = typeErr
		return
	}

	tagImpl, tagErr := newTag(fieldType.Tag.Get("orm"))
	if tagErr != nil {
		err = tagErr
		return
	}

	valueImpl, valueErr := newValue(fieldValue)
	if valueErr != nil {
		err = valueErr
		return
	}

	field := &fieldImpl{}
	field.fieldIndex = idx
	field.fieldName = fieldType.Name
	field.fieldType = typeImpl
	field.fieldTag = tagImpl
	field.fieldValue = valueImpl

	err = field.verify()
	if err != nil {
		return
	}

	ret = field
	return
}
