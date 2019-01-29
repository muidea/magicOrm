package local

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicOrm/model"
)

// modelInfo single struct ret
type modelInfo struct {
	structType reflect.Type

	fields model.Fields

	modelCache model.Cache
}

func (s *modelInfo) GetName() string {
	return s.structType.Name()
}

// GetPkgPath GetPkgPath
func (s *modelInfo) GetPkgPath() string {
	return s.structType.PkgPath()
}

// GetFields GetFields
func (s *modelInfo) GetFields() *model.Fields {
	return &s.fields
}

// SetFieldValue SetFieldValue
func (s *modelInfo) SetFieldValue(idx int, val reflect.Value) (err error) {
	for _, field := range s.fields {
		if field.GetIndex() == idx {
			err = field.SetValue(val)
			return
		}
	}

	return
}

// UpdateFieldValue UpdateFieldValue
func (s *modelInfo) UpdateFieldValue(name string, val reflect.Value) (err error) {
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
func (s *modelInfo) GetPrimaryField() model.Field {
	return s.fields.GetPrimaryField()
}

func (s *modelInfo) GetDependField() (ret []model.Field) {
	for _, field := range s.fields {
		fType := field.GetType()
		fDepend := fType.Depend()
		if fDepend != nil {
			ret = append(ret, field)
		}
	}

	return
}

func (s *modelInfo) Copy() model.Model {
	info := &modelInfo{structType: s.structType, fields: s.fields.Copy(), modelCache: s.modelCache}
	return info
}

func (s *modelInfo) Interface() reflect.Value {
	return reflect.New(s.structType)
}

// Dump Dump
func (s *modelInfo) Dump() {
	fmt.Print("modelInfo:\n")
	fmt.Printf("\tname:%s, pkgPath:%s\n", s.GetName(), s.GetPkgPath())

	primaryKey := s.fields.GetPrimaryField()
	if primaryKey != nil {
		fmt.Printf("primaryKey:\n")
		fmt.Printf("\t%s\n", primaryKey.Dump())
	}
	fmt.Print("fields:\n")
	s.fields.Dump()
}

// GetObjectModel GetObjectModel
func GetObjectModel(objPtr interface{}, cache model.Cache) (ret model.Model, err error) {
	ptrVal := reflect.ValueOf(objPtr)

	if ptrVal.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal obj type. must be a struct ptr")
		return
	}

	structVal := reflect.Indirect(ptrVal)
	structType := structVal.Type()

	ret, err = GetTypeModel(structType, cache)
	if err != nil {
		log.Printf("GetTypeModel failed, err:%s", err.Error())
		return
	}

	ret, err = GetValueModel(structVal, cache)
	if err != nil {
		log.Printf("GetValueModel failed, err:%s", err.Error())
		return
	}

	return
}

// GetTypeModel GetTypeModel
func GetTypeModel(structType reflect.Type, cache model.Cache) (ret model.Model, err error) {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal structType, type:%s", structType.String())
		return
	}

	info := cache.Fetch(structType.Name())
	if info != nil {
		ret = info
		return
	}

	modelInfo := &modelInfo{structType: structType, fields: make(model.Fields, 0), modelCache: cache}

	fieldNum := structType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := structType.Field(idx)
		fieldInfo, fieldErr := GetFieldInfo(idx, fieldType, nil)
		if fieldErr != nil {
			err = fieldErr
			log.Printf("getFieldInfo failed, name:%s, err:%s", fieldType.Name, err.Error())
			return
		}

		if fieldInfo != nil {
			modelInfo.fields.Append(fieldInfo)
		}
	}

	if len(modelInfo.fields) > 0 {
		cache.Put(modelInfo.GetName(), modelInfo)

		ret = modelInfo
		return
	}

	err = fmt.Errorf("no define orm field, struct name:%s", modelInfo.GetName())
	return
}

// GetValueModel GetValueModel
func GetValueModel(structVal reflect.Value, cache model.Cache) (ret model.Model, err error) {
	if structVal.Kind() == reflect.Ptr {
		if structVal.IsNil() {
			err = fmt.Errorf("can't get value from nil ptr")
			return
		}

		structVal = reflect.Indirect(structVal)
	}

	info := cache.Fetch(structVal.Type().Name())
	if info == nil {
		err = fmt.Errorf("can't get value modelInfo, valType:%s", structVal.Type().String())
		return
	}

	info = info.Copy()
	fieldNum := structVal.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		val := structVal.Field(idx)
		err = info.SetFieldValue(idx, val)
		if err != nil {
			log.Printf("SetFieldValue failed, err:%s", err.Error())
			return
		}
	}

	ret = info

	return
}

func getStructPrimaryKey(structVal reflect.Value) (ret model.Field, err error) {
	if structVal.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal value type, not struct, type:%s", structVal.Type().String())
		return
	}

	structType := structVal.Type()
	fieldNum := structType.NumField()
	for idx := 0; idx < fieldNum; {
		fieldType := structType.Field(idx)
		fieldVal := structVal.Field(idx)
		fieldInfo, fieldErr := GetFieldInfo(idx, fieldType, &fieldVal)
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

	err = fmt.Errorf("no found primary key. type:%s", structVal.Type().String())
	return
}
