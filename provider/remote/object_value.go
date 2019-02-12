package remote

import (
	"encoding/json"
	"log"
	"reflect"
)

// Value Value
type Value struct {
	TypeName string                 `json:"typeName"`
	PkgPath  string                 `json:"pkgPath"`
	Items    map[string]interface{} `json:"items"`
}

// Decode decode value
func (s *Value) Decode(data []byte) (err error) {
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
