package helper

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
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

func getSliceTypeDeclare(tType model.Type) (ret interface{}, err error) {
	switch tType.GetValue() {
	case util.TypeBooleanField:
	case util.TypeDateTimeField:
	case util.TypeFloatField, util.TypeDoubleField:
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
	case util.TypeStringField:
	default:
		err = fmt.Errorf("illegal type, type:%s", tType.GetName())
	}

	return
}

func (s *impl) decodeStringSlice(val string, tType model.Type) (ret model.Value, err error) {
	items := []string{}
	err = json.Unmarshal([]byte(val), &items)
	if err != nil {
		return
	}

	tVal, _ := tType.Interface()
	sliceVal := []interface{}{}
	for idx := range items {
		itemVal, itemErr := s.Decode(items[idx], tType.Elem())
		if itemErr != nil {
			err = itemErr
			return
		}

		sliceVal = append(sliceVal, itemVal.Get())
	}
	tVal.Set(sliceVal)

	ret = tVal
	return
}

func (s *impl) decodeReflectSlice(val reflect.Value, tType model.Type) (ret model.Value, err error) {
	tVal, _ := tType.Interface()
	sliceVal := []interface{}{}
	for idx := 0; idx < val.Len(); idx++ {
		itemVal, itemErr := s.Decode(val.Index(idx).Interface(), tType.Elem())
		if itemErr != nil {
			err = itemErr
			return
		}

		sliceVal = append(sliceVal, itemVal.Get())
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
