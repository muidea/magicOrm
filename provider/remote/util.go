package remote

import (
	"fmt"
	"reflect"
	"time"

	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

// GetValueModel GetValueModel
func GetValueModel(val Value, modelInfo model.Model) (err error) {
	if val.PkgPath != modelInfo.GetPkgPath() || val.TypeName != modelInfo.GetName() {
		err = fmt.Errorf("illegal value for modelInfo")
		return
	}

	for k, v := range val.Items {
		if v == nil {
			continue
		}

		err = modelInfo.UpdateFieldValue(k, reflect.ValueOf(v))
		if err != nil {
			return
		}
	}

	return
}

// GetItemValStr GetItemValStr
func GetItemValStr(item *Item, val interface{}) (ret string, err error) {
	rawVal := reflect.ValueOf(val)
	rawVal = reflect.Indirect(rawVal)
	switch item.Type.GetValue() {
	case util.TypeBitField:
		ret = fmt.Sprintf("%d", val.(int8))
	case util.TypePositiveBitField:
		ret = fmt.Sprintf("%d", val.(uint8))
	case util.TypeSmallIntegerField:
		ret = fmt.Sprintf("%d", val.(int16))
	case util.TypePositiveSmallIntegerField:
		ret = fmt.Sprintf("%d", val.(uint16))
	case util.TypeInteger32Field:
		ret = fmt.Sprintf("%d", val.(int32))
	case util.TypePositiveInteger32Field:
		ret = fmt.Sprintf("%d", val.(uint32))
	case util.TypeBigIntegerField:
		ret = fmt.Sprintf("%d", val.(int64))
	case util.TypePositiveBigIntegerField:
		ret = fmt.Sprintf("%d", val.(uint64))
	case util.TypeIntegerField:
		ret = fmt.Sprintf("%d", val.(int))
	case util.TypePositiveIntegerField:
		ret = fmt.Sprintf("%d", val.(uint))
	case util.TypeFloatField:
		ret = fmt.Sprintf("%f", val.(float32))
	case util.TypeDoubleField:
		ret = fmt.Sprintf("%f", val.(float64))
	case util.TypeBooleanField:
		if val.(bool) {
			ret = "1"
		} else {
			ret = "0"
		}
	case util.TypeStringField:
		ret = fmt.Sprintf("'%s'", val.(string))
	case util.TypeDateTimeField:
		ret = fmt.Sprintf("'%s'", val.(time.Time).Format("2006-01-02 15:04:05"))
	case util.TypeStructField:
		//ret = fmt.Sprintf("%d", rawVal.Interface().(int8))
	case util.TypeSliceField:
		//ret = fmt.Sprintf("%d", rawVal.Interface().(int8))
	default:
		err = fmt.Errorf("unsupport value type. type:%d", item.Type.Value)
	}
	return
}

// GetInfoValueStr GetInfoValueStr
func GetInfoValueStr(model Object, val interface{}) (ret string, err error) {
	//rawVal := reflect.ValueOf(val)
	//rawVal = reflect.Indirect(rawVal)

	return
}
