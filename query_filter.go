package orm

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"muidea.com/magicCommon/foundation/util"
	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/local"
	"muidea.com/magicOrm/model"
	ormutil "muidea.com/magicOrm/util"
)

func getBasicValStr(value reflect.Value) (ret string, err error) {
	switch value.Kind() {
	case reflect.Slice, reflect.Struct:
		err = fmt.Errorf("illegal basic type, type:%s", value.Type().String())
	case reflect.Bool:
		if value.Bool() {
			ret = "1"
		} else {
			ret = "0"
		}
	case reflect.String:
		ret = fmt.Sprintf("'%v'", value.Interface())
	default:
		ret = fmt.Sprintf("%v", value.Interface())
	}

	return
}

func getStructValStr(value reflect.Value) (ret string, err error) {
	switch value.Kind() {
	case reflect.Struct:
		if value.Type().String() == "time.Time" {
			ret = value.Interface().(time.Time).Format("2006-01-02 15:04:05")
			ret = fmt.Sprintf("'%s'", ret)
		} else {
			ret, err = local.GetModelValueStr(value)
		}
	default:
		err = fmt.Errorf("illegal struct type, type:%s", value.Type().String())
	}

	return
}

func getSliceValStr(value reflect.Value) (ret string, err error) {
	valSlice := []string{}
	pos := value.Len()
	for idx := 0; idx < pos; {
		sv := value.Index(idx)
		sv = reflect.Indirect(sv)
		strVal := ""
		switch sv.Kind() {
		case reflect.Slice:
			err = fmt.Errorf("illegal slice type, type:%s", value.Type().String())
		case reflect.Struct:
			strVal, err = getStructValStr(sv)
		default:
			strVal, err = getBasicValStr(sv)
		}

		if err != nil {
			return
		}

		valSlice = append(valSlice, strVal)
		idx++
	}

	ret = strings.Join(valSlice, ",")
	return
}

type filterItem struct {
	filterFun func(name, value string) string
	value     reflect.Value
}

func (s *filterItem) Verify(fType model.FieldType) (err error) {
	valType := s.value.Type()
	if valType.Kind() == reflect.Ptr {
		valType = valType.Elem()
	}
	if valType.Kind() == reflect.Slice {
		valType = valType.Elem()
	}

	fieldType := fType.Type()
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}
	if fieldType.Kind() == reflect.Slice {
		fieldType = fieldType.Elem()
	}

	if valType.Kind() != fieldType.Kind() {
		err = fmt.Errorf("illegal filter value, value type:%s, field type:%s", valType.String(), fieldType.String())
	}

	return
}

func (s *filterItem) FilterStr(name string) (ret string, err error) {
	fValue := ""
	switch s.value.Kind() {
	case reflect.Slice:
		fValue, err = getSliceValStr(s.value)
	case reflect.Struct:
		fValue, err = getStructValStr(s.value)
	default:
		fValue, err = getBasicValStr(s.value)
	}
	if err != nil {
		return
	}

	ret = s.filterFun(name, fValue)
	return
}

// queryFilter queryFilter
type queryFilter struct {
	params     map[string]model.FilterItem
	pageFilter *util.PageFilter
}

func (s *queryFilter) Equle(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormutil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if ormutil.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: builder.EquleOpr, value: qv}
	return
}

func (s *queryFilter) NotEqule(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormutil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if ormutil.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: builder.NotEquleOpr, value: qv}
	return
}

func (s *queryFilter) Below(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormutil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ormutil.IsBasicType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: builder.BelowOpr, value: qv}
	return
}

func (s *queryFilter) Above(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormutil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ormutil.IsBasicType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: builder.AboveOpr, value: qv}
	return
}

func (s *queryFilter) In(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormutil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ormutil.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: builder.InOpr, value: qv}
	return
}

func (s *queryFilter) NotIn(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormutil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ormutil.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: builder.NotInOpr, value: qv}
	return
}

func (s *queryFilter) Like(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	if qv.Kind() != reflect.String {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: builder.LikeOpr, value: qv}
	return
}

func (s *queryFilter) PageFilter(filter *util.PageFilter) {
	s.pageFilter = filter
}

func (s *queryFilter) Items() map[string]model.FilterItem {
	return s.params
}

func (s *queryFilter) Pagination() (limit, offset int, paging bool) {
	paging = false
	if s.pageFilter == nil {
		return
	}

	paging = true
	limit = s.pageFilter.PageSize
	offset = s.pageFilter.PageSize * (s.pageFilter.PageNum - 1)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 100
	}

	return
}
