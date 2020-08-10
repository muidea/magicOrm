package remote

import (
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

// SetFieldValue SetFieldValue
func (s *Object) SetFieldValue(idx int, val reflect.Value) (err error) {
	for _, item := range s.Items {
		if item.Index == idx {
			err = item.SetValue(val)
			if err != nil {
				log.Errorf("set field value failed, object name:%s, err:%s", s.Name, err.Error())
			}

			return
		}
	}

	err = fmt.Errorf("invalid field idx:%d", idx)
	return
}

// UpdateFieldValue UpdateFieldValue
func (s *Object) UpdateFieldValue(name string, val reflect.Value) (err error) {
	for _, item := range s.Items {
		if item.Name == name {
			err = item.UpdateValue(val)
			if err != nil {
				log.Errorf("update field value failed, object name:%s, err:%s", s.Name, err.Error())
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
func (s *Object) Interface() (ret reflect.Value) {
	val := &ObjectValue{Name: s.Name, PkgPath: s.PkgPath, Items: []*ItemValue{}}

	for _, v := range s.Items {
		val.Items = append(val.Items, v.Interface())
	}

	ret = reflect.ValueOf(val).Elem()
	return
}

// Copy Copy
func (s *Object) Copy() (ret *Object) {
	obj := &Object{Name: s.Name, PkgPath: s.PkgPath, Items: []*Item{}}
	for _, val := range s.Items {
		obj.Items = append(obj.Items, &Item{Index: val.Index, Name: val.Name, Tag: val.Tag.Copy(), Type: val.Type.Copy()})
	}

	ret = obj
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

	typeImpl, typeErr := newType(entityType)
	if typeErr != nil {
		err = fmt.Errorf("illegal obj type, must be a struct obj, type:%s", entityType.String())
		return
	}
	if typeImpl.GetValue() != util.TypeStructField {
		err = fmt.Errorf("illegal obj type, must be a struct obj, type:%s", entityType.String())
		return
	}

	ret = &Object{}
	//!! must be String, not Name
	ret.Name = entityType.String()
	ret.PkgPath = entityType.PkgPath()
	ret.Items = []*Item{}

	fieldNum := entityType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		field := entityType.Field(idx)
		fType := field.Type
		fTag := field.Tag.Get("orm")

		itemTag, itemErr := newTag(fTag)
		if itemErr != nil {
			err = itemErr
			return
		}

		itemType, itemErr := newType(fType)
		if itemErr != nil {
			err = itemErr
			return
		}

		fItem := &Item{Index: idx, Name: field.Name, Tag: itemTag, Type: itemType}

		ret.Items = append(ret.Items, fItem)
	}

	return
}
