package local

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicOrm/model"
)

// modelInfo single struct ret
type modelInfo struct {
	modelType reflect.Type

	fields model.Fields

	modelCache model.Cache
}

func (s *modelInfo) GetName() string {
	return s.modelType.Name()
}

// GetPkgPath GetPkgPath
func (s *modelInfo) GetPkgPath() string {
	return s.modelType.PkgPath()
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
	info := &modelInfo{modelType: s.modelType, fields: s.fields.Copy(), modelCache: s.modelCache}
	return info
}

func (s *modelInfo) Interface() reflect.Value {
	return reflect.New(s.modelType)
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

	modelVal := reflect.Indirect(ptrVal)

	ret, err = GetValueModel(modelVal, cache)
	if err != nil {
		log.Printf("GetValueModel failed, err:%s", err.Error())
		return
	}

	return
}

// GetTypeModel GetTypeModel
func GetTypeModel(modelType reflect.Type, cache model.Cache) (ret model.Model, err error) {
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal modelType, type:%s", modelType.String())
		return
	}

	info := cache.Fetch(modelType.Name())
	if info != nil {
		ret = info
		return
	}

	modelInfo := &modelInfo{modelType: modelType, fields: make(model.Fields, 0), modelCache: cache}

	fieldNum := modelType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := modelType.Field(idx)
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

		dependFields := modelInfo.GetDependField()
		for _, val := range dependFields {
			fType := val.GetType()
			fDValue := fType.Depend()
			if fDValue != nil {
				_, fDErr := GetTypeModel(fDValue.Type(), cache)
				if fDErr != nil {
					err = fDErr
					return
				}
			}
		}

		ret = modelInfo
		return
	}

	err = fmt.Errorf("no define orm field, struct name:%s", modelInfo.GetName())
	return
}

// GetValueModel GetValueModel
func GetValueModel(modelVal reflect.Value, cache model.Cache) (ret model.Model, err error) {
	if modelVal.Kind() == reflect.Ptr {
		if modelVal.IsNil() {
			err = fmt.Errorf("can't get value from nil ptr")
			return
		}

		modelVal = reflect.Indirect(modelVal)
	}

	info := cache.Fetch(modelVal.Type().Name())
	if info == nil {
		info, err = GetTypeModel(modelVal.Type(), cache)
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

	err = fmt.Errorf("no found primary key. type:%s", modelVal.Type().String())
	return
}
