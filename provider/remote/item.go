package remote

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

type Field struct {
	Index int    `json:"index"`
	Name  string `json:"name"`

	Type  *TypeImpl `json:"type"`
	Spec  *SpecImpl `json:"spec"`
	value *valueImpl
}

func (s *Field) GetIndex() (ret int) {
	return s.Index
}

func (s *Field) GetName() string {
	return s.Name
}

func (s *Field) GetType() (ret model.Type) {
	if s.Type != nil {
		ret = s.Type
	}
	return
}

func (s *Field) GetSpec() (ret model.Spec) {
	if s.Spec != nil {
		ret = s.Spec
	}

	return
}

func (s *Field) GetValue() (ret model.Value) {
	if s.value != nil {
		ret = s.value
		return
	}

	ret = s.Type.Interface()
	return
}

func (s *Field) IsPrimary() bool {
	return s.Spec.IsPrimaryKey()
}

func (s *Field) SetValue(val model.Value) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("SetValue failed, unexpect field:%v, err:%v", s.Name, err)
		}
	}()

	if s.value != nil {
		err = s.value.Set(val.Get())
		if err != nil {
			log.Errorf("set field value failed, name:%s, err:%s", s.Name, err.Error())
		}
		return
	}

	initVal := s.Type.Interface()
	initVal.Set(val.Get())
	s.value = &valueImpl{value: initVal.Get()}
	return
}

func (s *Field) copy() (ret model.Field) {
	return &Field{Index: s.Index, Name: s.Name, Spec: s.Spec, Type: s.Type, value: s.value}
}

func (s *Field) verify() (err error) {
	val := s.Type.GetValue()
	if s.Spec.IsAutoIncrement() {
		switch val {
		case util.TypeBooleanValue,
			util.TypeStringValue,
			util.TypeDateTimeValue,
			util.TypeFloatValue,
			util.TypeDoubleValue,
			util.TypeStructValue,
			util.TypeSliceValue:
			return fmt.Errorf("illegal auto_increment field type, type:%s", s.Type.dump())
		default:
		}
	}

	if s.Spec.IsPrimaryKey() {
		switch val {
		case util.TypeStructValue, util.TypeSliceValue:
			return fmt.Errorf("illegal primary key field type, type:%s", s.Type.dump())
		default:
		}
	}

	if s.value == nil || s.value.IsNil() {
		return nil
	}

	return s.value.verify()
}

func (s *Field) dump() string {
	str := fmt.Sprintf("index:%d,name:%s,type:[%s],spec:[%s]", s.Index, s.Name, s.Type.dump(), s.Spec.dump())
	if s.value != nil {
		str = fmt.Sprintf("%s,value:%v", str, s.value.Interface())
	}

	return str
}

func getFieldName(fieldType reflect.StructField) (ret string, err error) {
	specPtr, specErr := newSpec(fieldType.Tag)
	if specErr != nil {
		err = specErr
		return
	}

	fieldName := fieldType.Name
	if specPtr.GetFieldName() != "" {
		fieldName = specPtr.GetFieldName()
	}

	ret = fieldName
	return
}

func getItemInfo(idx int, fieldType reflect.StructField) (ret *Field, err error) {
	typeImpl, typeErr := newType(fieldType.Type)
	if typeErr != nil {
		err = typeErr
		return
	}

	specImpl, specErr := newSpec(fieldType.Tag)
	if specErr != nil {
		err = specErr
		return
	}

	initVal := typeImpl.Interface()

	item := &Field{}
	item.Index = idx
	item.Name = fieldType.Name
	if specImpl.GetFieldName() != "" {
		item.Name = specImpl.GetFieldName()
	}
	item.Type = typeImpl
	item.Spec = specImpl
	item.value = newValue(initVal.Get())

	ret = item
	return
}

func compareItem(l, r *Field) bool {
	if l.Index != r.Index {
		return false
	}
	if l.Name != r.Name {
		return false
	}

	if !compareType(l.Type, r.Type) {
		return false
	}
	if !compareSpec(l.Spec, r.Spec) {
		return false
	}

	return true
}
