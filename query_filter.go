package orm

import (
	"fmt"
	"reflect"

	"muidea.com/magicCommon/foundation/util"
	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/provider"
	ormutil "muidea.com/magicOrm/util"
)

type filterItem struct {
	filterFun     func(name, value string) string
	value         reflect.Value
	modelProvider provider.Provider
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
	fValue, fErr := s.modelProvider.GetValueStr(s.value)
	if fErr != nil {
		err = fErr
		return
	}

	ret = s.filterFun(name, fValue)
	return
}

// queryFilter queryFilter
type queryFilter struct {
	params        map[string]model.FilterItem
	pageFilter    *util.PageFilter
	modelProvider provider.Provider
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

	s.params[key] = &filterItem{filterFun: builder.EquleOpr, value: qv, modelProvider: s.modelProvider}
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

	s.params[key] = &filterItem{filterFun: builder.NotEquleOpr, value: qv, modelProvider: s.modelProvider}
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

	s.params[key] = &filterItem{filterFun: builder.BelowOpr, value: qv, modelProvider: s.modelProvider}
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

	s.params[key] = &filterItem{filterFun: builder.AboveOpr, value: qv, modelProvider: s.modelProvider}
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

	s.params[key] = &filterItem{filterFun: builder.InOpr, value: qv, modelProvider: s.modelProvider}
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

	s.params[key] = &filterItem{filterFun: builder.NotInOpr, value: qv, modelProvider: s.modelProvider}
	return
}

func (s *queryFilter) Like(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	if qv.Kind() != reflect.String {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: builder.LikeOpr, value: qv, modelProvider: s.modelProvider}
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
