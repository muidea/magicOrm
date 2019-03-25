package orm

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
	ormutil "github.com/muidea/magicOrm/util"
)

type filterValue struct {
	filterValue reflect.Value
}

func newFilterValue(val reflect.Value) (ret model.Value, err error) {
	if val.Kind() == reflect.Invalid {
		err = fmt.Errorf("illegal filter value")
		return
	}

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			err = fmt.Errorf("nil filter value")
			return
		}
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	_, qvErr := ormutil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}

	ret = &filterValue{filterValue: val}
	return
}

func (s *filterValue) IsNil() (ret bool) {
	if s.filterValue.Kind() == reflect.Invalid {
		return true
	}

	if s.filterValue.Kind() == reflect.Ptr {
		return s.filterValue.IsNil()
	}

	return false
}

func (s *filterValue) Set(val reflect.Value) (err error) {
	if val.Kind() == reflect.Invalid {
		return
	}
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
	}

	if s.filterValue.Kind() == reflect.Invalid {
		s.filterValue = val
		return
	}

	valTypeName := val.Type().String()
	expectTypeName := s.filterValue.Type().String()
	if expectTypeName != valTypeName {
		err = fmt.Errorf("illegal value type, type:%s, expect:%s", expectTypeName, valTypeName)
		return
	}

	s.filterValue.Set(val)
	return
}

func (s *filterValue) Update(val reflect.Value) (err error) {
	if val.Kind() == reflect.Invalid {
		return
	}
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
	}

	if s.filterValue.Kind() == reflect.Invalid {
		s.filterValue = val
		return
	}

	valTypeName := val.Type().String()
	expectTypeName := s.filterValue.Type().String()
	if expectTypeName != valTypeName {
		err = fmt.Errorf("illegal value type, type:%s, expect:%s", expectTypeName, valTypeName)
		return
	}

	s.filterValue.Set(val)
	return
}

func (s *filterValue) Get() (ret reflect.Value) {
	ret = s.filterValue

	return
}

type filterItem struct {
	filterFun     func(name, value string) string
	value         reflect.Value
	modelProvider provider.Provider
}

func (s *filterItem) verify(fType model.Type) (err error) {
	valType := s.value.Type()
	if valType.Kind() == reflect.Ptr {
		valType = valType.Elem()
	}
	if valType.Kind() == reflect.Slice {
		valType = valType.Elem()
	}

	fieldType := fType.GetType()
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

func (s *filterItem) FilterStr(name string, fType model.Type) (ret string, err error) {
	err = s.verify(fType)
	if err != nil {
		return
	}

	log.Printf("name:%s, type:%s", name, fType.GetType().String())
	fModel, fErr := s.modelProvider.GetTypeModel(fType)
	if fErr != nil {
		err = fErr
		return
	}
	if fModel != nil {
		fType = fType.Elem()
	}

	filterStr := ""
	if s.value.Kind() != reflect.Slice {
		filterVal, filterErr := newFilterValue(s.value)
		if filterErr != nil {
			err = filterErr
			return
		}

		fVal, fErr := s.modelProvider.GetValueStr(fType, filterVal)
		if fErr != nil {
			err = fErr
			return
		}

		filterStr = fVal
	} else {
		itemArray := []string{}
		for idx := 0; idx < s.value.Len(); idx++ {
			itemVal, itemErr := newFilterValue(s.value.Index(idx))
			if itemErr != nil {
				err = itemErr
				return
			}

			itemStr, itemErr := s.modelProvider.GetValueStr(fType, itemVal)
			if itemErr != nil {
				err = itemErr
				return
			}

			itemArray = append(itemArray, itemStr)
		}

		filterStr = strings.Join(itemArray, ",")
	}

	ret = s.filterFun(name, filterStr)
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
