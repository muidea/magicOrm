package local

import (
	"fmt"
	"log"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// modelImpl single model
type modelImpl struct {
	modelType reflect.Type
	fields    []*fieldImpl
}

func (s *modelImpl) GetName() string {
	return s.modelType.String()
}

// GetPkgPath GetPkgPath
func (s *modelImpl) GetPkgPath() string {
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

	err = fmt.Errorf("out of index, index:%d", idx)
	return
}

// UpdateFieldValue UpdateFieldValue
func (s *modelImpl) UpdateFieldValue(name string, val reflect.Value) (err error) {
	for _, field := range s.fields {
		if field.GetName() == name {
			err = field.UpdateValue(val)
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

func (s *modelImpl) Interface() reflect.Value {
	retVal := reflect.New(s.modelType).Elem()

	for _, val := range s.fields {
		tType := val.GetType()
		tVal := val.GetValue()
		if tType.IsPtrType() && !tVal.IsNil() {
			retVal.FieldByName(val.GetName()).Set(tType.Interface())
		}
	}

	return retVal
}

func (s *modelImpl) Copy() *modelImpl {
	modelInfo := &modelImpl{modelType: s.modelType, fields: []*fieldImpl{}}
	for _, field := range s.fields {
		modelInfo.fields = append(modelInfo.fields, field.Copy())
	}

	return modelInfo
}

// Dump Dump
func (s *modelImpl) Dump() (ret string) {
	ret = fmt.Sprintf("\nmodelImpl:\n")
	ret = fmt.Sprintf("%s\tname:%s, pkgPath:%s\n", ret, s.GetName(), s.GetPkgPath())

	ret = fmt.Sprintf("%sfields:\n", ret)
	for _, field := range s.fields {
		ret = fmt.Sprintf("%s\t%s\n", ret, field.Dump())
	}

	return
}

func getTypeModel(entityType reflect.Type) (ret model.Model, err error) {
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	hasPrimaryKey := false
	modelImpl := &modelImpl{modelType: entityType, fields: make([]*fieldImpl, 0)}
	fieldNum := entityType.NumField()
	var fieldValue reflect.Value
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := entityType.Field(idx)
		fieldInfo, fieldErr := getFieldInfo(idx, fieldType, fieldValue)
		if fieldErr != nil {
			err = fieldErr
			log.Printf("getFieldInfo failed, field idx:%d, field name:%s, struct name:%s, err:%s", idx, fieldType.Name, modelImpl.GetName(), err.Error())
			return
		}

		if fieldInfo.IsPrimary() {
			if hasPrimaryKey {
				err = fmt.Errorf("duplicate primary key field, struct name:%s", modelImpl.GetName())
				return
			}

			hasPrimaryKey = true
		}

		if fieldInfo != nil {
			modelImpl.fields = append(modelImpl.fields, fieldInfo)
		}
	}

	if len(modelImpl.fields) == 0 {
		err = fmt.Errorf("no define orm field, struct name:%s", modelImpl.GetName())
		return
	}
	if !hasPrimaryKey {
		err = fmt.Errorf("no define primary key field, struct name:%s", modelImpl.GetName())
		return
	}

	ret = modelImpl
	return
}

// getValueModel getValueModel
func getValueModel(modelVal reflect.Value) (ret *modelImpl, err error) {
	hasPrimaryKey := false
	modelVal = reflect.Indirect(modelVal)
	entityType := modelVal.Type()
	modelImpl := &modelImpl{modelType: entityType, fields: make([]*fieldImpl, 0)}
	fieldNum := entityType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldVal := modelVal.Field(idx)
		fieldType := entityType.Field(idx)
		fieldInfo, fieldErr := getFieldInfo(idx, fieldType, fieldVal)
		if fieldErr != nil {
			err = fieldErr
			log.Printf("getFieldInfo failed, field idx:%d, field name:%s, struct name:%s, err:%s", idx, fieldType.Name, modelImpl.GetName(), err.Error())
			return
		}

		if fieldInfo.IsPrimary() {
			if hasPrimaryKey {
				err = fmt.Errorf("duplicate primary key field, struct name:%s", modelImpl.GetName())
				return
			}

			hasPrimaryKey = true
		}

		if fieldInfo != nil {
			modelImpl.fields = append(modelImpl.fields, fieldInfo)
		}
	}

	if len(modelImpl.fields) == 0 {
		err = fmt.Errorf("no define orm field, struct name:%s", modelImpl.GetName())
		return
	}
	if !hasPrimaryKey {
		err = fmt.Errorf("no define primary key field, struct name:%s", modelImpl.GetName())
		return
	}

	ret = modelImpl
	return
}
