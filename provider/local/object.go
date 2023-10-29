package local

import (
	"fmt"
	"path"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

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
	for _, sf := range s.fields {
		ret = append(ret, sf)
	}

	return
}

func (s *objectImpl) SetFieldValue(name string, val model.Value) {
	for _, sf := range s.fields {
		if sf.GetName() == name {
			sf.SetValue(val)
			return
		}
	}

	return
}

func (s *objectImpl) SetPrimaryFieldValue(val model.Value) {
	for _, sf := range s.fields {
		if sf.IsPrimaryKey() {
			sf.SetValue(val)
			return
		}
	}

	return
}

func (s *objectImpl) GetPrimaryField() (ret model.Field) {
	for _, sf := range s.fields {
		if sf.IsPrimaryKey() {
			ret = sf
			return
		}
	}

	return
}

func (s *objectImpl) GetField(name string) (ret model.Field) {
	for _, sf := range s.fields {
		if sf.GetName() == name {
			ret = sf
			return
		}
	}

	return
}

func (s *objectImpl) Interface(ptrValue bool, viewSpec model.ViewDeclare) (ret interface{}) {
	retVal := reflect.New(s.objectType).Elem()

	for _, sf := range s.fields {
		if viewSpec > 0 {
			if sf.specPtr.EnableView(viewSpec) {
				fVal, _ := sf.typePtr.Interface(nil)
				val := fVal.Get().(reflect.Value)
				retVal.Field(sf.GetIndex()).Set(val)
			}

			continue
		}

		fVal := sf.GetValue()
		if fVal.IsNil() {
			continue
		}

		val := fVal.Get().(reflect.Value)
		retVal.Field(sf.GetIndex()).Set(val)
	}

	if ptrValue {
		retVal = retVal.Addr()
	}

	ret = retVal.Interface()
	return
}

func (s *objectImpl) Copy() model.Model {
	objectPtr := &objectImpl{objectType: s.objectType, fields: []*field{}}
	for _, sf := range s.fields {
		objectPtr.fields = append(objectPtr.fields, sf.copy())
	}

	return objectPtr
}

func (s *objectImpl) Dump() (ret string) {
	ret = fmt.Sprintf("\nmodelImpl:\n")
	ret = fmt.Sprintf("%s\tname:%s, pkgPath:%s\n", ret, s.GetName(), s.GetPkgPath())

	ret = fmt.Sprintf("%s fields:\n", ret)
	for _, sf := range s.fields {
		ret = fmt.Sprintf("%s\t%s\n", ret, sf.dump())
	}

	return
}

func (s *objectImpl) verify() (err *cd.Result) {
	if s.GetName() == "" || s.GetPkgPath() == "" {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal object declare informain"))
		return
	}

	for _, sf := range s.fields {
		err = sf.verify()
		if err != nil {
			log.Errorf("verify field failed, idx:%d, name:%s, err:%s", sf.index, sf.name, err.Error())
			return
		}
	}

	return
}

func getTypeModel(entityType reflect.Type) (ret *objectImpl, err *cd.Result) {
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	typeImpl, typeErr := NewType(entityType)
	if typeErr != nil {
		err = typeErr
		log.Errorf("getTypeModel failed, err:%s", err.Error())
		return
	}
	if typeImpl.GetValue() != model.TypeStructValue {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal type, must be a struct entity, type:%s", entityType.String()))
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

		if tField.IsPrimaryKey() {
			if hasPrimaryKey {
				err = cd.NewError(cd.UnExpected, fmt.Sprintf("duplicate primary key field, field idx:%d,field name:%s, struct name:%s", idx, fieldInfo.Name, impl.GetName()))

				log.Errorf("getTypeModel failed, check primary key err:%s", err.Error())
				return
			}

			hasPrimaryKey = true
		}

		impl.fields = append(impl.fields, tField)
	}

	if len(impl.fields) == 0 {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("no define orm field, struct name:%s", impl.GetName()))
		log.Errorf("getTypeModel failed, check fields err:%s", err.Error())
		return
	}
	if !hasPrimaryKey {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("no define primary key field, struct name:%s", impl.GetName()))
		log.Errorf("getTypeModel failed, check primary key err:%s", err.Error())
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

func getValueModel(modelVal reflect.Value) (ret *objectImpl, err *cd.Result) {
	modelVal = reflect.Indirect(modelVal)
	entityType := modelVal.Type()
	typeImpl, typeErr := NewType(entityType)
	if typeErr != nil {
		err = typeErr
		log.Errorf("getValueModel failed, err:%s", err.Error())
		return
	}
	if typeImpl.GetValue() != model.TypeStructValue {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal type, must be a struct entity, type:%s", entityType.String()))
		log.Errorf("getValueModel failed, err:%s", err.Error())
		return
	}

	hasPrimaryKey := false
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

		if tField.IsPrimaryKey() {
			if hasPrimaryKey {
				err = cd.NewError(cd.UnExpected, fmt.Sprintf("duplicate primary key field, field idx:%d,field name:%s, struct name:%s", idx, fieldInfo.Name, impl.GetName()))
				log.Errorf("getValueModel failed, check primary key err:%s", err.Error())
				return
			}

			hasPrimaryKey = true
		}

		impl.fields = append(impl.fields, tField)
	}

	if len(impl.fields) == 0 {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("no define orm field, struct name:%s", impl.GetName()))
		log.Errorf("getValueModel failed, check fields err:%s", err.Error())
		return
	}
	if !hasPrimaryKey {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("no define primary key field, struct name:%s", impl.GetName()))
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
