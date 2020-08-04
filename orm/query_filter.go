package orm

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
	ormUtil "github.com/muidea/magicOrm/util"
)

type filterValue struct {
	filterValue reflect.Value
}

func newFilterValue(val reflect.Value) (ret model.Value, err error) {
	if ormUtil.IsNil(val) {
		err = fmt.Errorf("illegal filter value")
		return
	}

	qv := reflect.Indirect(val)
	_, qvErr := ormUtil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}

	ret = &filterValue{filterValue: val}
	return
}

func (s *filterValue) IsNil() (ret bool) {
	return ormUtil.IsNil(s.filterValue)
}

func (s *filterValue) Set(val reflect.Value) (err error) {
	if ormUtil.IsNil(val) {
		return
	}

	s.filterValue = val
	return
}

func (s *filterValue) Update(val reflect.Value) (err error) {
	if ormUtil.IsNil(val) {
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

func isNumber(vKind reflect.Kind) (ret bool) {
	switch vKind {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint,
		reflect.Float32, reflect.Float64:
		ret = true
	default:
		ret = false
	}

	return
}

func (s *filterItem) FilterStr(name string, fType model.Type) (ret string, err error) {
	fModel, fErr := s.modelProvider.GetTypeModel(fType)
	if fErr != nil {
		err = fErr
		return
	}
	if fModel != nil {
		fType = fType.Depend()
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
		var itemArray []string
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
	maskValue     interface{}
	pageFilter    *util.PageFilter
	sortFilter    *util.SortFilter
	modelProvider provider.Provider
}

func (s *queryFilter) Equal(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormUtil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if ormUtil.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: builder.EqualOpr, value: qv, modelProvider: s.modelProvider}
	return
}

func (s *queryFilter) NotEqual(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormUtil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if ormUtil.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: builder.NotEqualOpr, value: qv, modelProvider: s.modelProvider}
	return
}

func (s *queryFilter) Below(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormUtil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ormUtil.IsBasicType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: builder.BelowOpr, value: qv, modelProvider: s.modelProvider}
	return
}

func (s *queryFilter) Above(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormUtil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ormUtil.IsBasicType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: builder.AboveOpr, value: qv, modelProvider: s.modelProvider}
	return
}

func (s *queryFilter) In(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormUtil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ormUtil.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}
	if qv.Len() > 0 {
		s.params[key] = &filterItem{filterFun: builder.InOpr, value: qv, modelProvider: s.modelProvider}
	}

	return
}

func (s *queryFilter) NotIn(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormUtil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ormUtil.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	if qv.Len() > 0 {
		s.params[key] = &filterItem{filterFun: builder.NotInOpr, value: qv, modelProvider: s.modelProvider}
	}
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

func (s *queryFilter) ValueMask(val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormUtil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}

	if !ormUtil.IsStructType(qvType) {
		err = fmt.Errorf("illegal mask value")
		return
	}

	s.maskValue = val
	return
}

func (s *queryFilter) Page(filter *util.PageFilter) {
	s.pageFilter = filter
}

func (s *queryFilter) Sort(sorter *util.SortFilter) {
	s.sortFilter = sorter
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

func (s *queryFilter) MaskModel() (ret model.Model, err error) {
	if s.maskValue != nil {
		maskVal := reflect.New(reflect.TypeOf(s.maskValue))
		ret, err = s.modelProvider.GetEntityModel(maskVal.Interface())
	}

	return
}

func (s *queryFilter) Sorter() model.Sorter {
	if s.sortFilter == nil {
		return nil
	}

	return s.sortFilter
}
