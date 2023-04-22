package local

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicCommon/foundation/util"
	om "github.com/muidea/magicOrm/model"
	ou "github.com/muidea/magicOrm/util"
)

type filterItem struct {
	oprCode om.OprCode
	value   *valueImpl
}

func (s *filterItem) OprCode() om.OprCode {
	return s.oprCode
}

func (s *filterItem) OprValue() om.Value {
	return s.value
}

type filter struct {
	bindType   *typeImpl
	params     map[string]*filterItem
	maskValue  *valueImpl
	pageFilter *util.Pagination
	sortFilter *util.SortFilter
}

func NewFilter(typePtr *typeImpl) *filter {
	return &filter{bindType: typePtr, params: map[string]*filterItem{}}
}

func (s *filter) Equal(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ou.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if ou.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	//s.equalFilter = append(s.equalFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: om.Equal, value: newValue(qv)}
	return
}

func (s *filter) NotEqual(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ou.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if ou.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	//s.notEqualFilter = append(s.notEqualFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: om.NotEqual, value: newValue(qv)}
	return
}

func (s *filter) Below(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ou.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ou.IsBasicType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	//s.belowFilter = append(s.belowFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: om.Below, value: newValue(qv)}
	return
}

func (s *filter) Above(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ou.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ou.IsBasicType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	//s.aboveFilter = append(s.aboveFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: om.Above, value: newValue(qv)}
	return
}

func (s *filter) In(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ou.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ou.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	//s.inFilter = append(s.inFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: om.In, value: newValue(qv)}
	return
}

func (s *filter) NotIn(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ou.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ou.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	//s.notInFilter = append(s.notInFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: om.NotIn, value: newValue(qv)}
	return
}

func (s *filter) Like(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	if qv.Kind() != reflect.String {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	//s.likeFilter = append(s.likeFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: om.Like, value: newValue(qv)}
	return
}

func (s *filter) Page(pageFilter *util.Pagination) {
	s.pageFilter = pageFilter
}

func (s *filter) Sort(sorter *util.SortFilter) {
	s.sortFilter = sorter
}

func (s *filter) ValueMask(val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ou.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ou.IsStructType(qvType) {
		err = fmt.Errorf("illegal mask value, type:%s", qv.Type().String())
		return
	}

	bindType := s.bindType.getType().String()
	maskType := reflect.Indirect(qv).Type().String()
	if bindType != maskType {
		err = fmt.Errorf("miscmatch mask value, bindType:%v, maskType:%v", bindType, maskType)
		return
	}

	s.maskValue = newValue(qv)
	return
}

func (s *filter) GetFilterItem(key string) om.FilterItem {
	v, ok := s.params[key]
	if ok {
		return v
	}

	return nil
}

func (s *filter) Pagination() (limit, offset int, paging bool) {
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

func (s *filter) Sorter() om.Sorter {
	if s.sortFilter == nil {
		return nil
	}

	return s.sortFilter
}

func (s *filter) MaskModel() (ret om.Model) {
	maskVal := s.bindType.Interface()
	if s.maskValue != nil {
		maskVal = s.maskValue
	}

	objPtr, objErr := getValueModel(maskVal.Get())
	if objErr != nil {
		return
	}

	ret = objPtr
	return
}
