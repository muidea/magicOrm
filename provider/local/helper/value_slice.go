package helper

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// encodeSlice get slice value str
func (s *impl) encodeSlice(vVal model.Value, tType model.Type) (ret string, err error) {
	vals, valErr := s.elemDependValue(vVal)
	if valErr != nil {
		err = valErr
		return
	}
	items := []string{}
	for _, val := range vals {
		strVal, strErr := s.Encode(val, tType.Elem())
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

func (s *impl) decodeStringSlice(val string, tType model.Type) (ret model.Value, err error) {
	items := []string{}
	err = json.Unmarshal([]byte(val), &items)
	if err != nil {
		return
	}

	tVal, _ := tType.Interface()
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

	ret = tVal
	return
}

func (s *impl) decodeReflectSlice(val reflect.Value, tType model.Type) (ret model.Value, err error) {
	tVal, _ := tType.Interface()
	sliceVal := tVal.Get().(reflect.Value)
	sliceVal = reflect.Indirect(sliceVal)
	for idx := 0; idx < val.Len(); idx++ {
		itemVal, itemErr := s.Decode(val.Index(idx).Interface(), tType.Elem())
		if itemErr != nil {
			err = itemErr
			return
		}

		sliceVal = reflect.Append(sliceVal, itemVal.Get().(reflect.Value))
	}
	tVal.Set(sliceVal)

	ret = tVal
	return
}

// decodeSlice decode slice from string
func (s *impl) decodeSlice(val interface{}, tType model.Type) (ret model.Value, err error) {
	rVal := reflect.ValueOf(val)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}
	rVal = reflect.Indirect(rVal)
	switch rVal.Kind() {
	case reflect.String:
		ret, err = s.decodeStringSlice(rVal.String(), tType)
	case reflect.Slice:
		ret, err = s.decodeReflectSlice(rVal, tType)
	default:
		err = fmt.Errorf("illegal slice value, val:%v", val)
	}
	return
}
