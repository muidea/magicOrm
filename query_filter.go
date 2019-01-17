package orm

import (
	"fmt"
	"reflect"

	"muidea.com/magicCommon/foundation/util"
	"muidea.com/magicOrm/model"
	ormutil "muidea.com/magicOrm/util"
)

type filterItem struct {
	name      string
	filterFun func(name string, value queryValue) (string, error)
	value     reflect.Value
}

// queryFilter queryFilter
type queryFilter struct {
	params         map[string]filterItem
	pageFilter     *util.PageFilter
	modelInfoCache model.StructInfoCache
}

type queryValue interface {
	String() (string, error)
}

func newQueryValue(qv interface{}, cache model.StructInfoCache) (ret queryValue, err error) {
	val := reflect.Indirect(reflect.ValueOf(qv))
	fval, fErr := ormutil.GetTypeValueEnum(val.Type())
	if fErr != nil {
		err = fErr
		return
	}

	if ormutil.IsBasicType(fval) {
		ret = &basicValue{value: val}
		return
	}

	if ormutil.IsStructType(fval) {
		ret = &structValue{value: val, modelInfoCache: cache}
		return
	}

	if ormutil.IsSliceType(fval) {
		ret = &sliceValue{value: val, modelInfoCache: cache}
		return
	}

	err = fmt.Errorf("illegal query value, type:%s", val.Type().String())
	return
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

	s.params[key] = filterItem{name: key, filterFun: equleOpr, value: qv}
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

	s.params[key] = filterItem{name: key, filterFun: notEquleOpr, value: qv}
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

	s.params[key] = filterItem{name: key, filterFun: belowOpr, value: qv}
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

	s.params[key] = filterItem{name: key, filterFun: aboveOpr, value: qv}
	return
}

func (s *queryFilter) In(key string, val []interface{}) (err error) {
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

	s.params[key] = filterItem{name: key, filterFun: inOpr, value: qv}
	return
}

func (s *queryFilter) NotIn(key string, val []interface{}) (err error) {
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

	s.params[key] = filterItem{name: key, filterFun: notInOpr, value: qv}
	return
}

func (s *queryFilter) PageFilter(filter *util.PageFilter) {
	s.pageFilter = filter
}

func (s *queryFilter) Builder(structInfo model.StructInfo) (ret string, err error) {
	if structInfo == nil {
		return
	}

	fields := structInfo.GetFields()
	for _, field := range *fields {
		fType := field.GetFieldType()
		fDepend, _ := fType.Depend()
		if fDepend != nil {
			continue
		}

		filterItem, ok := s.params[field.GetFieldName()]
		if !ok {
			continue
		}

		basicValue := &basicValue{value: filterItem.value}
		strVal, strErr := basicValue.String()
		if strErr != nil {
			err = strErr
			return
		}

		strVal, strErr = filterItem.filterFun(field.GetFieldName(), basicValue)
		if strErr != nil {
			err = strErr
			return
		}

		if ret == "" {
			ret = fmt.Sprintf("%s", strVal)
		} else {
			ret = fmt.Sprintf("%s AND %s", ret, strVal)
		}
	}

	return
}

func (s *queryFilter) buildRelation(structInfo model.StructInfo) (ret string, err error) {
	if structInfo == nil {
		return
	}

	fields := structInfo.GetFields()
	for _, field := range *fields {
		fType := field.GetFieldType()
		fDepend, _ := fType.Depend()
		if fDepend == nil {
			continue
		}

		_, ok := s.params[field.GetFieldName()]
		if !ok {
			continue
		}
	}

	return
}
