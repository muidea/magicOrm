package remote

import (
	"encoding/json"
	"fmt"
	"path"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	pu "github.com/muidea/magicOrm/provider/util"
)

type Object struct {
	Name    string   `json:"name"`
	PkgPath string   `json:"pkgPath"`
	IsPtr   bool     `json:"isPtr"`
	Fields  []*Field `json:"fields"`
}

func (s *Object) GetName() (ret string) {
	ret = s.Name
	return
}

func (s *Object) GetPkgPath() (ret string) {
	ret = s.PkgPath
	return
}

func (s *Object) GetPkgKey() string {
	return path.Join(s.GetPkgPath(), s.GetName())
}

func (s *Object) IsPtrValue() bool {
	return s.IsPtr
}

func (s *Object) GetFields() (ret model.Fields) {
	for _, val := range s.Fields {
		ret = append(ret, val)
	}

	return
}

func (s *Object) SetFieldValue(name string, val model.Value) (err error) {
	for _, item := range s.Fields {
		if item.Name == name {
			err = item.SetValue(val)
			if err != nil {
				log.Errorf("set field value failed, object name:%s, err:%s", s.Name, err.Error())
			}

			return
		}
	}

	err = fmt.Errorf("invalid field name:%s", name)
	return
}

func (s *Object) GetPrimaryField() (ret model.Field) {
	for _, v := range s.Fields {
		if v.IsPrimary() {
			ret = v
			return
		}
	}

	return
}

func (s *Object) GetField(name string) (ret model.Field) {
	for _, v := range s.Fields {
		if v.GetName() == name {
			ret = v
			return
		}
	}

	return
}

func (s *Object) itemInterface(valPtr *Field) (ret interface{}) {
	rVal := valPtr.value.Get()
	if !valPtr.Type.IsBasic() {
		rVal = rVal.Addr()
		if model.IsStructType(valPtr.Type.GetValue()) {
			objectVal := rVal.Interface().(*ObjectValue)
			if len(objectVal.Fields) > 0 {
				ret = objectVal
			}
		}
		if model.IsSliceType(valPtr.Type.GetValue()) {
			sliceObjectVal := rVal.Interface().(*SliceObjectValue)
			if len(sliceObjectVal.Values) > 0 {
				ret = sliceObjectVal
			}
		}

		return
	}

	ret = rVal.Interface()
	return
}

// Interface object value
func (s *Object) Interface(ptrValue bool) (ret interface{}) {
	objVal := &ObjectValue{Name: s.Name, PkgPath: s.PkgPath, Fields: []*FieldValue{}}

	for _, v := range s.Fields {
		if v.value.IsNil() {
			objVal.Fields = append(objVal.Fields, &FieldValue{Name: v.Name})
			continue
		}

		interfaceVal := s.itemInterface(v)
		objVal.Fields = append(objVal.Fields, &FieldValue{Name: v.Name, Value: interfaceVal})
	}

	if ptrValue {
		ret = objVal
		return
	}

	ret = *objVal
	return
}

func (s *Object) Copy() (ret model.Model) {
	obj := &Object{Name: s.Name, PkgPath: s.PkgPath, Fields: []*Field{}}
	for _, val := range s.Fields {
		item := &Field{Index: val.Index, Name: val.Name, Type: val.Type.copy()}
		if val.Spec != nil {
			item.Spec = val.Spec.copy()
		}
		if val.value != nil {
			item.value = val.value.Copy()
		} else {
			initVal := val.Type.Interface()
			item.value = pu.NewValue(initVal.Get())
		}

		obj.Fields = append(obj.Fields, item)
	}

	ret = obj
	return
}

func (s *Object) Dump() (ret string) {
	ret = fmt.Sprintf("\nmodelImpl:\n")
	ret = fmt.Sprintf("%s\tname:%s, pkgPath:%s\n", ret, s.GetName(), s.GetPkgPath())

	ret = fmt.Sprintf("%sfields:\n", ret)
	for _, field := range s.Fields {
		ret = fmt.Sprintf("%s\t%s\n", ret, field.dump())
	}

	return
}

func (s *Object) Verify() (err error) {
	if s.Name == "" || s.PkgPath == "" {
		err = fmt.Errorf("illegal object declare informain")
		return
	}

	for _, val := range s.Fields {
		err = val.verify()
		if err != nil {
			log.Errorf("Verify field failed, idx:%d, name:%s, err:%s", val.Index, val.Name, err.Error())
			return
		}
	}
	return
}

// GetObject GetObject
func GetObject(entity interface{}) (ret *Object, err error) {
	entityType := reflect.ValueOf(entity).Type()
	ret, err = type2Object(entityType)
	if err != nil {
		log.Errorf("type2Object failed, raw type:%s, err:%s", entityType.String(), err.Error())
	}

	return
}

// type2Object type2Object
func type2Object(entityType reflect.Type) (ret *Object, err error) {
	isPtr := false
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
		isPtr = true
	}
	if entityType.Kind() == reflect.Interface {
		entityType = entityType.Elem()
	}
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
		isPtr = true
	}

	typeImpl, typeErr := newType(entityType)
	if typeErr != nil {
		err = fmt.Errorf("illegal entity type, must be a struct obj, type:%s", entityType.String())
		return
	}
	if !model.IsStructType(typeImpl.GetValue()) {
		err = fmt.Errorf("illegal obj type, must be a struct obj, type:%s", entityType.String())
		return
	}

	impl := &Object{}
	impl.Name = entityType.Name()
	impl.PkgPath = entityType.PkgPath()
	impl.IsPtr = isPtr
	impl.Fields = []*Field{}

	hasPrimaryKey := false
	fieldNum := entityType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := entityType.Field(idx)
		fItem, fErr := getItemInfo(idx, fieldType)
		if fErr != nil {
			err = fErr
			return
		}
		if fItem.IsPrimary() {
			if hasPrimaryKey {
				err = fmt.Errorf("duplicate primary key field, field idx:%d,field name:%s, struct name:%s", idx, fieldType.Name, impl.GetName())
				return
			}

			hasPrimaryKey = true
		}

		impl.Fields = append(impl.Fields, fItem)
	}

	if len(impl.Fields) == 0 {
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

func EncodeObject(objPtr *Object) (ret []byte, err error) {
	ret, err = json.Marshal(objPtr)
	return
}

func DecodeObject(data []byte) (ret *Object, err error) {
	objPtr := &Object{}
	err = json.Unmarshal(data, objPtr)
	if err != nil {
		return
	}

	ret = objPtr
	return
}

func compareObject(l, r *Object) bool {
	if l.Name != r.Name {
		return false
	}

	if l.PkgPath != r.PkgPath {
		return false
	}

	if len(l.Fields) != len(r.Fields) {
		return false
	}

	for idx := 0; idx < len(l.Fields); idx++ {
		lVal := l.Fields[idx]
		rVal := r.Fields[idx]
		if !compareItem(lVal, rVal) {
			return false
		}
	}

	return true
}
