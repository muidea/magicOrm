package local

import (
	"fmt"
	"path"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
)

type objectImpl struct {
	objectType reflect.Type
	fields     []*field
}

func (s *objectImpl) GetName() string {
	return s.objectType.Name()
}

func (s *objectImpl) GetPkgPath() string {
	return s.objectType.PkgPath()
}

func (s *objectImpl) GetDescription() string {
	return ""
}

func (s *objectImpl) GetPkgKey() string {
	return path.Join(s.GetPkgPath(), s.GetName())
}

func (s *objectImpl) GetFields() (ret model.Fields) {
	for _, field := range s.fields {
		ret = append(ret, field)
	}

	return
}

func (s *objectImpl) SetFieldValue(name string, val model.Value) (err error) {
	for _, field := range s.fields {
		if field.GetName() == name {
			err = field.SetValue(val)
			return
		}
	}

	err = fmt.Errorf("illegal field,model name:%s, field name:%s", s.GetName(), name)
	return
}

func (s *objectImpl) GetPrimaryField() (ret model.Field) {
	for _, field := range s.fields {
		if field.IsPrimary() {
			ret = field
			return
		}
	}

	return
}

func (s *objectImpl) GetField(name string) (ret model.Field) {
	for _, field := range s.fields {
		if field.GetName() == name {
			ret = field
			return
		}
	}

	return
}

func (s *objectImpl) Interface(ptrValue bool) (ret interface{}) {
	retVal := reflect.New(s.objectType).Elem()

	for _, field := range s.fields {
		tVal := field.GetValue()
		if tVal.IsNil() {
			continue
		}

		val := tVal.Get()
		if !field.GetType().IsPtrType() {
			val = reflect.Indirect(val)
		}
		retVal.Field(field.GetIndex()).Set(val)
	}

	if ptrValue {
		retVal = retVal.Addr()
	}

	ret = retVal.Interface()
	return
}

func (s *objectImpl) Copy() model.Model {
	objectPtr := &objectImpl{objectType: s.objectType, fields: []*field{}}
	for _, field := range s.fields {
		objectPtr.fields = append(objectPtr.fields, field.copy())
	}

	return objectPtr
}

func (s *objectImpl) Dump() (ret string) {
	ret = fmt.Sprintf("\nmodelImpl:\n")
	ret = fmt.Sprintf("%s\tname:%s, pkgPath:%s\n", ret, s.GetName(), s.GetPkgPath())

	ret = fmt.Sprintf("%sfields:\n", ret)
	for _, field := range s.fields {
		ret = fmt.Sprintf("%s\t%s\n", ret, field.dump())
	}

	return
}

func (s *objectImpl) verify() (err error) {
	if s.GetName() == "" || s.GetPkgPath() == "" {
		err = fmt.Errorf("illegal object declare informain")
		return
	}

	for _, val := range s.fields {
		err = val.verify()
		if err != nil {
			log.Errorf("verify field failed, idx:%d, name:%s, err:%s", val.index, val.name, err.Error())
			return
		}
	}

	return
}

func getTypeModel(entityType reflect.Type) (ret *objectImpl, err error) {
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	typeImpl, typeErr := newType(entityType)
	if typeErr != nil {
		err = fmt.Errorf("illegal type, must be a struct entity, type:%s", entityType.String())
		log.Errorf("getTypeModel failed, err:%s", err.Error())
		return
	}
	if typeImpl.GetValue() != model.TypeStructValue {
		err = fmt.Errorf("illegal type, must be a struct entity, type:%s", entityType.String())
		log.Errorf("getTypeModel failed, err:%s", err.Error())
		return
	}

	hasPrimaryKey := false
	impl := &objectImpl{objectType: entityType, fields: make([]*field, 0)}
	fieldNum := entityType.NumField()
	var fieldValue reflect.Value
	for idx := 0; idx < fieldNum; idx++ {
		fieldInfo := entityType.Field(idx)
		tField, tErr := getFieldInfo(idx, fieldInfo, fieldValue)
		if tErr != nil {
			err = tErr
			log.Errorf("getTypeModel failed, field idx:%d, field name:%s, struct name:%s, err:%s", idx, fieldInfo.Name, impl.GetName(), err.Error())
			return
		}

		if tField.IsPrimary() {
			if hasPrimaryKey {
				err = fmt.Errorf("duplicate primary key field, field idx:%d,field name:%s, struct name:%s", idx, fieldInfo.Name, impl.GetName())

				log.Errorf("getTypeModel failed, check primary key err:%s", err.Error())
				return
			}

			hasPrimaryKey = true
		}

		impl.fields = append(impl.fields, tField)
	}

	if len(impl.fields) == 0 {
		err = fmt.Errorf("no define orm field, struct name:%s", impl.GetName())
		log.Errorf("getTypeModel failed, check fields err:%s", err.Error())
		return
	}
	if !hasPrimaryKey {
		err = fmt.Errorf("no define primary key field, struct name:%s", impl.GetName())
		log.Errorf("getTypeModel failed, check primary key err:%s", err.Error())
		return
	}

	ret = impl
	return
}

func getValueModel(modelVal reflect.Value) (ret *objectImpl, err error) {
	hasPrimaryKey := false
	modelVal = reflect.Indirect(modelVal)
	entityType := modelVal.Type()
	impl := &objectImpl{objectType: entityType, fields: []*field{}}
	fieldNum := entityType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldVal := modelVal.Field(idx)
		fieldInfo := entityType.Field(idx)
		tField, tErr := getFieldInfo(idx, fieldInfo, fieldVal)
		if tErr != nil {
			err = tErr
			log.Errorf("getValueModel failed, field idx:%d, field name:%s, struct name:%s, err:%s", idx, fieldInfo.Name, impl.GetName(), err.Error())
			return
		}

		if tField.IsPrimary() {
			if hasPrimaryKey {
				err = fmt.Errorf("duplicate primary key field, field idx:%d,field name:%s, struct name:%s", idx, fieldInfo.Name, impl.GetName())
				log.Errorf("getValueModel failed, check primary key err:%s", err.Error())
				return
			}

			hasPrimaryKey = true
		}

		if tField != nil {
			impl.fields = append(impl.fields, tField)
		}
	}

	if len(impl.fields) == 0 {
		err = fmt.Errorf("no define orm field, struct name:%s", impl.GetName())
		log.Errorf("getValueModel failed, check fields err:%s", err.Error())
		return
	}
	if !hasPrimaryKey {
		err = fmt.Errorf("no define primary key field, struct name:%s", impl.GetName())
		log.Errorf("getValueModel failed, check primary key err:%s", err.Error())
		return
	}

	err = impl.verify()
	if err != nil {
		log.Errorf("verify model failed, err:%s", err.Error())
		return
	}

	ret = impl
	return
}
