package local

import (
	"fmt"
	"log"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// modelImpl single model
type modelImpl struct {
	modelType reflect.Type
	fields    []*field
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
func (s *modelImpl) SetFieldValue(idx int, val model.Value) (err error) {
	for _, field := range s.fields {
		if field.GetIndex() == idx {
			err = field.SetValue(val)
			return
		}
	}

	err = fmt.Errorf("out of index, index:%d", idx)
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

func (s *modelImpl) Interface() (ret model.Value) {
	retVal := reflect.New(s.modelType).Elem()

	for _, field := range s.fields {
		tVal := field.GetValue()
		if tVal.IsNil() {
			continue
		}

		retVal.Field(field.GetIndex()).Set(tVal.Get().(reflect.Value))
	}

	ret = newValue(retVal)
	return
}

func (s *modelImpl) Copy() model.Model {
	modelInfo := &modelImpl{modelType: s.modelType, fields: []*field{}}
	for _, field := range s.fields {
		modelInfo.fields = append(modelInfo.fields, field.copy())
	}

	return modelInfo
}

// Dump Dump
func (s *modelImpl) Dump() (ret string) {
	ret = fmt.Sprintf("\nmodelImpl:\n")
	ret = fmt.Sprintf("%s\tname:%s, pkgPath:%s\n", ret, s.GetName(), s.GetPkgPath())

	ret = fmt.Sprintf("%sfields:\n", ret)
	for _, field := range s.fields {
		ret = fmt.Sprintf("%s\t%s\n", ret, field.dump())
	}

	return
}

func getTypeModel(entityType reflect.Type) (ret *modelImpl, err error) {
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	typeImpl, typeErr := newType(entityType)
	if typeErr != nil {
		err = fmt.Errorf("illegal type, must be a struct entity, type:%s", entityType.String())
		return
	}
	if typeImpl.GetValue() != util.TypeStructField {
		err = fmt.Errorf("illegal type, must be a struct entity, type:%s", entityType.String())
		return
	}

	hasPrimaryKey := false
	impl := &modelImpl{modelType: entityType, fields: make([]*field, 0)}
	fieldNum := entityType.NumField()
	var fieldValue reflect.Value
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := entityType.Field(idx)
		fieldInfo, fieldErr := getFieldInfo(idx, fieldType, fieldValue)
		if fieldErr != nil {
			err = fieldErr
			log.Printf("getFieldInfo failed, field idx:%d, field name:%s, struct name:%s, err:%s", idx, fieldType.Name, impl.GetName(), err.Error())
			return
		}

		if fieldInfo.IsPrimary() {
			if hasPrimaryKey {
				err = fmt.Errorf("duplicate primary key field, struct name:%s", impl.GetName())
				return
			}

			hasPrimaryKey = true
		}

		impl.fields = append(impl.fields, fieldInfo)
	}

	if len(impl.fields) == 0 {
		err = fmt.Errorf("no define orm field, struct name:%s", impl.GetName())
		return
	}
	if !hasPrimaryKey {
		err = fmt.Errorf("no define primary key field, struct name:%s", impl.GetName())
		return
	}

	ret = impl
	return
}

// getValueModel getValueModel
func getValueModel(modelVal reflect.Value) (ret *modelImpl, err error) {
	hasPrimaryKey := false
	modelVal = reflect.Indirect(modelVal)
	entityType := modelVal.Type()
	impl := &modelImpl{modelType: entityType, fields: make([]*field, 0)}
	fieldNum := entityType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldVal := modelVal.Field(idx)
		fieldType := entityType.Field(idx)
		fieldInfo, fieldErr := getFieldInfo(idx, fieldType, fieldVal)
		if fieldErr != nil {
			err = fieldErr
			log.Printf("getFieldInfo failed, field idx:%d, field name:%s, struct name:%s, err:%s", idx, fieldType.Name, impl.GetName(), err.Error())
			return
		}

		if fieldInfo.IsPrimary() {
			if hasPrimaryKey {
				err = fmt.Errorf("duplicate primary key field, struct name:%s", impl.GetName())
				return
			}

			hasPrimaryKey = true
		}

		if fieldInfo != nil {
			impl.fields = append(impl.fields, fieldInfo)
		}
	}

	if len(impl.fields) == 0 {
		err = fmt.Errorf("no define orm field, struct name:%s", impl.GetName())
		return
	}
	if !hasPrimaryKey {
		err = fmt.Errorf("no define primary key field, struct name:%s", impl.GetName())
		return
	}

	ret = impl
	return
}
