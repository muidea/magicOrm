package remote

import (
	"encoding/json"
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// Object Object
type Object struct {
	Name    string  `json:"name"`
	PkgPath string  `json:"pkgPath"`
	Items   []*Item `json:"items"`
}

// GetName GetName
func (s *Object) GetName() (ret string) {
	ret = s.Name
	return
}

// GetPkgPath GetPkgPath
func (s *Object) GetPkgPath() (ret string) {
	ret = s.PkgPath
	return
}

// GetFields GetFields
func (s *Object) GetFields() (ret model.Fields) {
	for _, val := range s.Items {
		ret = append(ret, val)
	}

	return
}

// UpdateFieldValue UpdateFieldValue
func (s *Object) SetFieldValue(name string, val model.Value) (err error) {
	for _, item := range s.Items {
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

// GetPrimaryField GetPrimaryField
func (s *Object) GetPrimaryField() (ret model.Field) {
	for _, v := range s.Items {
		if v.IsPrimary() {
			ret = v
			return
		}
	}

	return
}

// Interface Interface
func (s *Object) Interface() (ret model.Value) {
	val := &ObjectValue{Name: s.Name, PkgPath: s.PkgPath, Items: []*ItemValue{}}

	for _, v := range s.Items {
		val.Items = append(val.Items, &ItemValue{Name: v.Name, Value: v.value.Get()})
	}

	ret = newValue(val)
	return
}

// Copy Copy
func (s *Object) Copy() (ret model.Model) {
	obj := &Object{Name: s.Name, PkgPath: s.PkgPath, Items: []*Item{}}
	for _, val := range s.Items {
		item := &Item{Index: val.Index, Name: val.Name, Tag: val.Tag.copy(), Type: val.Type.copy()}
		if val.value != nil {
			item.value = val.value.copy()
		}

		obj.Items = append(obj.Items, item)
	}

	ret = obj
	return
}

func (s *Object) Dump() (ret string) {
	ret = fmt.Sprintf("\nmodelImpl:\n")
	ret = fmt.Sprintf("%s\tname:%s, pkgPath:%s\n", ret, s.GetName(), s.GetPkgPath())

	ret = fmt.Sprintf("%sfields:\n", ret)
	for _, field := range s.Items {
		ret = fmt.Sprintf("%s\t%s\n", ret, field.dump())
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
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	if entityType.Kind() == reflect.Interface {
		entityType = entityType.Elem()
	}
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	typeImpl, typeErr := newType(entityType)
	if typeErr != nil {
		err = fmt.Errorf("illegal obj type, must be a struct obj, type:%s", entityType.String())
		return
	}
	if util.IsStructType(typeImpl.GetValue()) {
		err = fmt.Errorf("illegal obj type, must be a struct obj, type:%s", entityType.String())
		return
	}

	impl := &Object{}
	//!! must be String, not Name
	impl.Name = entityType.String()
	impl.PkgPath = entityType.PkgPath()
	impl.Items = []*Item{}

	hasPrimaryKey := false
	fieldNum := entityType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldInfo := entityType.Field(idx)
		fItem, fErr := getItemInfo(idx, fieldInfo)
		if fErr != nil {
			err = fErr
			return
		}
		if fItem.IsPrimary() {
			if hasPrimaryKey {
				err = fmt.Errorf("duplicate primary key field, field idx:%d,field name:%s, struct name:%s", idx, fieldInfo.Name, impl.GetName())
				return
			}

			hasPrimaryKey = true
		}

		impl.Items = append(impl.Items, fItem)
	}

	if len(impl.Items) == 0 {
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

	if len(l.Items) != len(r.Items) {
		return false
	}

	for idx := 0; idx < len(l.Items); idx++ {
		lVal := l.Items[idx]
		rVal := r.Items[idx]
		if !compareItem(lVal, rVal) {
			return false
		}
	}

	return true
}
