package local

import (
	"fmt"
	"reflect"
)

// encodeStructValue get struct value str
func encodeStructValue(val reflect.Value, cache Cache) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	modelImpl, modelErr := getValueModel(rawVal, cache)
	if modelErr != nil {
		err = modelErr
		return
	}

	pk := modelImpl.GetPrimaryField()
	if pk == nil {
		err = fmt.Errorf("no define primary field")
		return
	}

	ret, err = getValueStr(pk.GetType(), pk.GetValue(), cache)

	return
}
