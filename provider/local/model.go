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
			err = field.SetValue(val)
			return
		}
	}

	return
}

// UpdateFieldValue UpdateFieldValue
func (s *modelImpl) UpdateFieldValue(name string, val reflect.Value) (err error) {
	for _, field := range s.fields {
		if field.GetName() == name {
			err = field.SetValue(val)
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

func (s *modelImpl) GetDependField() (ret []model.Field, err error) {
	for _, field := range s.fields {
		fType := field.GetType()
		fDepend, fErr := fType.GetDepend()
		if fErr != nil {
			err = fErr
			return
		}

		if fDepend != nil {
			ret = append(ret, field)
		}
	}

	return
}

func (s *modelImpl) IsPtr() bool {
	return s.modelType.Kind() == reflect.Ptr
}

func (s *modelImpl) Copy() model.Model {
	info := &modelImpl{modelType: s.modelType, fields: []*fieldImpl{}}
	for _, field := range s.fields {
		info.fields = append(info.fields, field.Copy())
	}

	return info
}

func (s *modelImpl) Interface() reflect.Value {
	if s.modelType.Kind() == reflect.Ptr {
		return reflect.New(s.modelType.Elem())
	}

	return reflect.New(s.modelType)
}

// Dump Dump
func (s *modelImpl) Dump() (ret string) {
	ret = fmt.Sprintf("modelImpl:\n")
	ret = fmt.Sprintf("%s\tname:%s, pkgPath:%s\n", ret, s.GetName(), s.GetPkgPath())

	primaryKey := s.GetPrimaryField()
	if primaryKey != nil {
		ret = fmt.Sprintf("%sprimaryKey:\n", ret)
		ret = fmt.Sprintf("%s\t%s\n", ret, primaryKey.Dump())
	}
	ret = fmt.Sprint("%sfields:\n", ret)
	for _, field := range s.fields {
		ret = fmt.Sprintf("%s%s\n", ret, field.Dump())
	}
	return
}

// GetObjectModel GetObjectModel
func GetObjectModel(objPtr interface{}, cache Cache) (ret model.Model, err error) {
	ptrVal := reflect.ValueOf(objPtr)

	if ptrVal.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal obj type. must be a struct ptr")
		return
	}

	modelVal := reflect.Indirect(ptrVal)

	ret, err = GetValueModel(modelVal, cache)
	if err != nil {
		log.Printf("GetValueModel failed, err:%s", err.Error())
		return
	}

	return
}

func getTypeModel(modelType reflect.Type, cache Cache) (ret model.Model, err error) {
	rawType := modelType
	if modelType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
	}

	if rawType.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal modelType, type:%s", rawType.String())
		return
	}

	info := cache.Fetch(rawType.Name())
	if info != nil {
		ret = info
		return
	}

	modelImpl := &modelImpl{modelType: modelType, fields: make([]*fieldImpl, 0)}

	fieldNum := rawType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := rawType.Field(idx)
		fieldInfo, fieldErr := getFieldInfo(idx, fieldType, nil)
		if fieldErr != nil {
			err = fieldErr
			log.Printf("getFieldInfo failed, name:%s, err:%s", fieldType.Name, err.Error())
			return
		}

		if fieldInfo != nil {
			modelImpl.fields = append(modelImpl.fields, fieldInfo)
		}
	}

	if len(modelImpl.fields) > 0 {
		cache.Put(modelImpl.GetName(), modelImpl)
		ret = modelImpl
		return
	}

	err = fmt.Errorf("no define orm field, struct name:%s", modelImpl.GetName())
	return
}

// GetValueModel GetValueModel
func GetValueModel(modelVal reflect.Value, cache Cache) (ret model.Model, err error) {
	rawVal := modelVal
	if modelVal.Kind() == reflect.Ptr {
		if modelVal.IsNil() {
			err = fmt.Errorf("can't get value from nil ptr")
			return
		}

		modelVal = reflect.Indirect(modelVal)
	}

	info := cache.Fetch(modelVal.Type().Name())
	if info == nil {
		info, err = getTypeModel(rawVal.Type(), cache)
		if err != nil {
			return
		}
	}

	info = info.Copy()
	fieldNum := modelVal.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		val := modelVal.Field(idx)
		err = info.SetFieldValue(idx, val)
		if err != nil {
			log.Printf("SetFieldValue failed, err:%s", err.Error())
			return
		}
	}

	ret = info

	return
}

func getStructPrimaryKey(modelVal reflect.Value) (ret model.Field, err error) {
	if modelVal.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal value type, not struct, type:%s", modelVal.Type().String())
		return
	}

	modelType := modelVal.Type()
	fieldNum := modelType.NumField()
	for idx := 0; idx < fieldNum; {
		fieldType := modelType.Field(idx)
		fieldVal := modelVal.Field(idx)
		fieldInfo, fieldErr := getFieldInfo(idx, fieldType, &fieldVal)
		if fieldErr != nil {
			err = fieldErr
			return
		}

		fTag := fieldInfo.GetTag()
		if fTag.IsPrimaryKey() {
			ret = fieldInfo
			return
		}

		idx++
	}

	err = fmt.Errorf("no found primary key. type:%s", modelVal.Type().String())
	return
}
