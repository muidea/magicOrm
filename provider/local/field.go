package local

import (
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// field single field impl
type field struct {
	index int
	name  string

	typePtr  *TypeImpl
	specPtr  *SpecImpl
	valuePtr *ValueImpl
}

func (s *field) GetIndex() int {
	return s.index
}

func (s *field) GetName() string {
	return s.name
}

func (s *field) GetShowName() string {
	return s.name
}

func (s *field) GetDescription() string {
	return ""
}

func (s *field) GetType() (ret model.Type) {
	ret = s.typePtr
	return
}

func (s *field) GetSpec() (ret model.Spec) {
	if s.specPtr != nil {
		ret = s.specPtr
		return
	}

	ret = &emptySpec
	return
}

func (s *field) GetValue() (ret model.Value) {
	ret = s.valuePtr
	return
}

func (s *field) SetValue(val any) *cd.Error {
	return s.valuePtr.Set(val)
}

func (s *field) GetSliceValue() []model.Value {
	if !model.IsSlice(s.typePtr) || !s.valuePtr.IsValid() {
		return nil
	}

	return s.valuePtr.UnpackValue()
}

func (s *field) AppendSliceValue(val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.UnExpected, "field append slice value is nil")
		return
	}

	if !model.IsSlice(s.typePtr) {
		err = cd.NewError(cd.UnExpected, "field is not slice")
		return
	}

	err = s.valuePtr.Append(reflect.ValueOf(val))
	return
}

func (s *field) Reset() {
	s.valuePtr.reset(model.IsAssignedField(s))
}

func getFieldInfo(idx int, fieldType reflect.StructField, fieldValue reflect.Value, viewSpec model.ViewDeclare) (ret *field, err *cd.Error) {
	var typePtr *TypeImpl
	var specPtr *SpecImpl
	var valuePtr *ValueImpl

	typePtr, err = NewType(fieldType.Type)
	if err != nil {
		return
	}

	specPtr, err = NewSpec(fieldType.Tag)
	if err != nil {
		return
	}

	fieldPtr := &field{
		index: idx,
		name:  fieldType.Name,
	}

	if specPtr.GetFieldName() != "" {
		fieldPtr.name = specPtr.GetFieldName()
	}

	valuePtr = NewValue(fieldValue)

	switch viewSpec {
	case model.MetaView:
		if !typePtr.IsPtrType() {
			valuePtr.reset(true)
		} else {
			valuePtr.reset(false)
		}
	case model.DetailView, model.LiteView:
		if !specPtr.EnableView(viewSpec) {
			// 如果spec未定义，则重置该value，不进行初始化
			valuePtr.reset(false)
		} else if !valuePtr.IsValid() {
			// 如果spec定义，但是value无效，则重置该value，并进行初始化
			valuePtr.reset(true)
		}
	case model.OriginView:
		//  do nothing
	default:
		log.Warnf("fieldName:%s,unknown view spec:%v", fieldPtr.name, viewSpec)
	}

	fieldPtr.typePtr = typePtr
	fieldPtr.specPtr = specPtr
	fieldPtr.valuePtr = valuePtr

	ret = fieldPtr
	return
}
