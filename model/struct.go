package model

import (
	"fmt"
	"log"
	"reflect"
)

// Model Model
type Model interface {
	GetName() string
	GetPkgPath() string
	GetFields() *Fields
	SetFieldValue(idx int, val reflect.Value) error
	UpdateFieldValue(name string, val reflect.Value) error
	GetPrimaryField() FieldInfo
	GetDependField() []FieldInfo
	Copy() Model
	Interface() reflect.Value
	Dump()
}

// modelInfo single struct ret
type modelInfo struct {
	structType reflect.Type

	fields Fields

	structInfoCache Cache
}

func (s *modelInfo) GetName() string {
	return s.structType.Name()
}

// GetPkgPath GetPkgPath
func (s *modelInfo) GetPkgPath() string {
	return s.structType.PkgPath()
}

// GetFields GetFields
func (s *modelInfo) GetFields() *Fields {
	return &s.fields
}

// SetFieldValue SetFieldValue
func (s *modelInfo) SetFieldValue(idx int, val reflect.Value) (err error) {
	for _, field := range s.fields {
		if field.GetFieldIndex() == idx {
			err = field.SetFieldValue(val)
			return
		}
	}

	return
}

// UpdateFieldValue UpdateFieldValue
func (s *modelInfo) UpdateFieldValue(name string, val reflect.Value) (err error) {
	for _, field := range s.fields {
		if field.GetFieldName() == name {
			err = field.SetFieldValue(val)
			return
		}
	}

	err = fmt.Errorf("no found field, name:%s", name)
	return
}

// GetPrimaryField GetPrimaryField
func (s *modelInfo) GetPrimaryField() FieldInfo {
	return s.fields.GetPrimaryField()
}

func (s *modelInfo) GetDependField() (ret []FieldInfo) {
	for _, field := range s.fields {
		fType := field.GetFieldType()
		fDepend, _ := fType.Depend()
		if fDepend != nil {
			ret = append(ret, field)
		}
	}

	return
}

func (s *modelInfo) Copy() Model {
	info := &modelInfo{structType: s.structType, fields: s.fields.Copy(), structInfoCache: s.structInfoCache}
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

// GetObjectStructInfo GetObjectStructInfo
func GetObjectStructInfo(objPtr interface{}, cache Cache) (ret Model, err error) {
	ptrVal := reflect.ValueOf(objPtr)

	if ptrVal.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal obj type. must be a struct ptr")
		return
	}

	structVal := reflect.Indirect(ptrVal)
	structType := structVal.Type()

	ret, err = GetStructInfo(structType, cache)
	if err != nil {
		log.Printf("GetStructInfo failed, err:%s", err.Error())
		return
	}

	ret, err = GetStructValue(structVal, cache)
	if err != nil {
		log.Printf("GetStructValue failed, err:%s", err.Error())
		return
	}

	return
}

// GetStructInfo GetStructInfo
func GetStructInfo(structType reflect.Type, cache Cache) (ret Model, err error) {
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

	modelInfo := &modelInfo{structType: structType, fields: make(Fields, 0), structInfoCache: cache}

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

// GetStructValue GetStructValue
func GetStructValue(structVal reflect.Value, cache Cache) (ret Model, err error) {
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

func getStructPrimaryKey(structVal reflect.Value) (ret FieldInfo, err error) {
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

		fTag := fieldInfo.GetFieldTag()
		if fTag.IsPrimaryKey() {
			ret = fieldInfo
			return
		}

		idx++
	}

	err = fmt.Errorf("no found primary key. type:%s", structVal.Type().String())
	return
}
