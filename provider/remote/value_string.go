package remote

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// EncodeStringValue get string value str
func EncodeStringValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%s", rawVal.String())

	return
}

func DecodeStringValue(val string, vType model.Type) (ret reflect.Value, err error) {
	if vType.GetType().Kind() != reflect.String {
		err = fmt.Errorf("illegal value type")
		return
	}

	ret = reflect.Indirect(vType.Interface())
	ret.SetString(val)

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}
