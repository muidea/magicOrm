package remote

import (
	"encoding/json"
	"log"
	"reflect"
)

// ObjectValue Object Value
type ObjectValue struct {
	TypeName string                 `json:"typeName"`
	PkgPath  string                 `json:"pkgPath"`
	Items    map[string]interface{} `json:"items"`
}

// GetName get object name
func (s *ObjectValue) GetName() string {
	return s.TypeName
}

// GetPkgPath get pkgpath
func (s *ObjectValue) GetPkgPath() string {
	return s.PkgPath
}

// GetObjectValue get object value
func GetObjectValue(obj interface{}) (ret *ObjectValue, err error) {
	ret = &ObjectValue{Items: map[string]interface{}{}}
	objValue := reflect.Indirect(reflect.ValueOf(obj))

	objType := objValue.Type()
	ret.TypeName = objType.Name()
	ret.PkgPath = objType.PkgPath()

	fieldNum := objValue.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := objType.Field(idx)
		fieldValue := objValue.Field(idx)

		if fieldType.Type.Kind() == reflect.Struct {

			continue
		}
		ret.Items[fieldType.Name] = fieldValue.Interface()
	}

	return
}

// Decode decode value
func (s *ObjectValue) Decode(data []byte) (err error) {
	err = json.Unmarshal(data, s)
	return
}

// EncodeObject encode Object
func EncodeObject(obj interface{}) (ret []byte, err error) {
	vType := reflect.TypeOf(obj)
	if vType.Kind() == reflect.Ptr {
		vType = vType.Elem()
	}

	objVal := map[string]interface{}{}
	objVal["typeName"] = vType.Name()
	objVal["pkgPath"] = vType.PkgPath()
	objVal["items"] = obj

	ret, err = json.Marshal(&objVal)
	if err != nil {
		log.Printf("EncodeObject failed, marshal exception, err:%s", err.Error())
	}

	return
}
