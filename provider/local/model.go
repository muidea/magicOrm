package local

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicOrm/model"
)

// modelImpl single model
type modelImpl struct {
	modelType reflect.Type

	fields []*fieldImpl
}

func (s *modelImpl) GetName() string {
	if s.modelType.Kind() == reflect.Ptr {
		return s.modelType.Elem().Name()
	}

	return s.modelType.Name()
}

// GetPkgPath GetPkgPath
func (s *modelImpl) GetPkgPath() string {
	if s.modelType.Kind() == reflect.Ptr {
		return s.modelType.Elem().PkgPath()
	}

	return s.modelType.PkgPath()
}

// GetFields GetFields
func (s *modelImpl) GetFields() (ret model.Fields) {
	for _, field := range s.fields {
		ret = append(ret, field)
	}

	return
}

// SetFieldValue SetFieldValue
func (s *modelImpl) SetFieldValue(idx int, val reflect.Value) (err error) {
	for _, field := range s.fields {
		if field.GetIndex() == idx {
			fv := field.GetValue()
			err = fv.Set(val)
			return
		}
	}

	err = fmt.Errorf("out of index, index:%d", idx)
	return
}

// UpdateFieldValue UpdateFieldValue
func (s *modelImpl) UpdateFieldValue(name string, val reflect.Value) (err error) {
	for _, field := range s.fields {
		if field.GetName() == name {
			fv := field.GetValue()
			err = fv.Set(val)
			return
		}
	}

	err = fmt.Errorf("no found field, name:%s", name)
	return
}

// GetPrimaryField GetPrimaryField
func (s *modelImpl) GetPrimaryField() (ret model.Field) {
	for _, field := range s.fields {
		if field.IsPrimary() {
			ret = field
			return
		}
	}

	return
}

func (s *modelImpl) IsPtr() bool {
	return s.modelType.Kind() == reflect.Ptr
}

func (s *modelImpl) Interface() reflect.Value {
	rawType := s.modelType
	if rawType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
	}

	return reflect.New(rawType).Elem()
}

func (s *modelImpl) Copy() *modelImpl {
	modelInfo := &modelImpl{modelType: s.modelType, fields: []*fieldImpl{}}
	for _, field := range s.fields {
		modelInfo.fields = append(modelInfo.fields, field.Copy())
	}

	return modelInfo
}

// Dump Dump
func (s *modelImpl) Dump(cache Cache) (ret string) {
	ret = fmt.Sprintf("\nmodelImpl:\n")
	ret = fmt.Sprintf("%s\tname:%s, pkgPath:%s\n", ret, s.GetName(), s.GetPkgPath())

	ret = fmt.Sprintf("%sfields:\n", ret)
	for _, field := range s.fields {
		ret = fmt.Sprintf("%s\t%s\n", ret, field.Dump(cache))
	}

	log.Print(ret)

	return
}

// getObjectModel GetObjectModel
func getObjectModel(objPtr interface{}, cache Cache) (ret *modelImpl, err error) {
	ptrVal := reflect.ValueOf(objPtr)

	if ptrVal.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal obj type. must be a struct ptr")
		return
	}

	modelVal := reflect.Indirect(ptrVal)

	ret, err = getValueModel(modelVal, cache)
	if err != nil {
		log.Printf("getValueModel failed, err:%s", err.Error())
		return
	}

	return
}

// getTypeModel getTypeModel
func getTypeModel(modelType reflect.Type, cache Cache) (ret *modelImpl, err error) {
	rawType := modelType
	if modelType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
	}

	if rawType.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal modelType, type:%s", rawType.String())
		return
	}
	if rawType.String() == "time.Time" {
		err = fmt.Errorf("illegal modelType, type:%s", rawType.String())
		return
	}

	modelInfo := cache.Fetch(rawType.Name())
	if modelInfo != nil {
		ret = modelInfo
		return
	}

	modelImpl := &modelImpl{modelType: modelType, fields: make([]*fieldImpl, 0)}

	fieldNum := rawType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := rawType.Field(idx)
		fieldInfo, fieldErr := getFieldInfo(idx, fieldType)
		if fieldErr != nil {
			err = fieldErr
			log.Printf("getFieldInfo failed, name:%s, err:%s", fieldType.Name, err.Error())
			return
		}

		if fieldInfo != nil {
			modelImpl.fields = append(modelImpl.fields, fieldInfo)
		}
	}

	if len(modelImpl.fields) == 0 {
		err = fmt.Errorf("no define orm field, struct name:%s", modelImpl.GetName())
		return
	}

	cache.Put(modelImpl.GetName(), modelImpl)

	ret = modelImpl
	return
}

// getValueModel getValueModel
func getValueModel(modelVal reflect.Value, cache Cache) (ret *modelImpl, err error) {
	rawVal := modelVal
	if modelVal.Kind() == reflect.Ptr {
		if modelVal.IsNil() {
			err = fmt.Errorf("can't get value model from nil ptr")
			return
		}

		modelVal = reflect.Indirect(modelVal)
	}

	modelInfo := cache.Fetch(modelVal.Type().Name())
	if modelInfo == nil {
		modelInfo, err = getTypeModel(rawVal.Type(), cache)
		if err != nil {
			return
		}
	}

	modelInfo = modelInfo.Copy()
	fieldNum := modelVal.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		val := modelVal.Field(idx)
		err = modelInfo.SetFieldValue(idx, val)
		if err != nil {
			log.Printf("SetFieldValue failed, err:%s", err.Error())
			return
		}
	}

	ret = modelInfo

	return
}

// getValueStr get value str
func getValueStr(vType model.Type, vVal model.Value, cache Cache) (ret string, err error) {
	rawType := vType.GetType()
	if rawType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
	}

	switch rawType.Kind() {
	case reflect.Bool:
		ret, err = getBoolValueStr(vVal.Get())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		ret, err = getIntValueStr(vVal.Get())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		ret, err = getUintValueStr(vVal.Get())
	case reflect.Float32, reflect.Float64:
		ret, err = getFloatValueStr(vVal.Get())
	case reflect.String:
		ret, err = getStringValueStr(vVal.Get())
	case reflect.Slice:
		ret, err = getSliceValueStr(vVal.Get())
	case reflect.Struct:
		ret, err = getDateTimeValueStr(vVal.Get())
	default:
		err = fmt.Errorf("illegal value kind, kind:%v", rawType.Kind())
	}
	return
}
