package helper

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// encodeSliceValue get slice value str
func (s *impl) encodeSliceValue(vVal model.Value, tType model.Type) (ret string, err error) {
	val := vVal.Get().(reflect.Value)
	val = reflect.Indirect(val)
	if val.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal slice value, type:%s", val.Type().String())
		return
	}

	items := []string{}
	for idx := 0; idx < val.Len(); idx++ {
		strVal, strErr := s.Encode(s.getValue(val.Index(idx)), tType.Elem())
		if strErr != nil {
			err = strErr
			return
		}

		items = append(items, strVal)
	}

	data, dataErr := json.Marshal(items)
	if dataErr != nil {
		err = dataErr
		return
	}

	ret = string(data)
	return
}

// decodeSliceValue decode slice from string
func (s *impl) decodeSliceValue(val string, tType model.Type) (ret model.Value, err error) {
	items := []string{}
	err = json.Unmarshal([]byte(val), &items)
	if err != nil {
		return
	}
	tVal := tType.Interface(nil)
	sliceVal := tVal.Get().(reflect.Value)
	sliceVal = reflect.Indirect(sliceVal)
	for idx := range items {
		itemVal, itemErr := s.Decode(items[idx], tType.Elem())
		if itemErr != nil {
			err = itemErr
			return
		}

		sliceVal = reflect.Append(sliceVal, itemVal.Get().(reflect.Value))
	}
	tVal.Set(sliceVal)
	if tType.IsPtrType() {
		tVal = tVal.Addr()
	}

	ret = tVal
	return
}
