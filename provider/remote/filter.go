package remote

import (
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
	om "github.com/muidea/magicOrm/model"
	pu "github.com/muidea/magicOrm/provider/util"
)

type filterItem struct {
	oprCode om.OprCode
	value   *pu.ValueImpl
}

func (s *filterItem) OprCode() om.OprCode {
	return s.oprCode
}

func (s *filterItem) OprValue() om.Value {
	return s.value
}

type ObjectFilter struct {
	Name           string           `json:"name"`
	PkgPath        string           `json:"pkgPath"`
	EqualFilter    []*FieldValue    `json:"equal"`
	NotEqualFilter []*FieldValue    `json:"noEqual"`
	BelowFilter    []*FieldValue    `json:"below"`
	AboveFilter    []*FieldValue    `json:"above"`
	InFilter       []*FieldValue    `json:"in"`
	NotInFilter    []*FieldValue    `json:"notIn"`
	LikeFilter     []*FieldValue    `json:"like"`
	MaskValue      *ObjectValue     `json:"maskValue"`
	PageFilter     *util.Pagination `json:"page"`
	SortFilter     *util.SortFilter `json:"sort"`

	bindObject *Object
}

func NewFilter(objectPtr *Object) *ObjectFilter {
	return &ObjectFilter{
		Name:           objectPtr.GetName(),
		PkgPath:        objectPtr.GetPkgPath(),
		EqualFilter:    []*FieldValue{},
		NotEqualFilter: []*FieldValue{},
		BelowFilter:    []*FieldValue{},
		AboveFilter:    []*FieldValue{},
		InFilter:       []*FieldValue{},
		NotInFilter:    []*FieldValue{},
		LikeFilter:     []*FieldValue{},
		bindObject:     objectPtr,
	}
}

func (s *ObjectFilter) GetName() string {
	return s.Name
}

func (s *ObjectFilter) GetPkgPath() string {
	return s.PkgPath
}

func (s *ObjectFilter) GetString(key string) (ret string, ok bool) {
	for _, item := range s.EqualFilter {
		if item.Name == key {
			ret, ok = (item.Value).(string)
			return
		}
	}

	return
}

func (s *ObjectFilter) GetInt(key string) (ret int, ok bool) {
	for _, item := range s.EqualFilter {
		if item.Name == key {
			val, vOK := (item.Value).(float64)
			if !vOK {
				return
			}

			ret = int(val)
			ok = true
			return
		}
	}

	return
}

func (s *ObjectFilter) Equal(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := pu.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if om.IsSliceType(qvType) {
		err = fmt.Errorf("equal failed, illegal value type, type:%s", qv.Type().String())
		return
	}

	if qvType == om.TypeDateTimeValue {
		val = qv.Interface().(time.Time).Format(util.CSTLayout)
	}

	if om.IsBasicType(qvType) {
		item := &FieldValue{Name: key, Value: val}
		s.EqualFilter = append(s.EqualFilter, item)
		return
	}

	if om.IsMapType(qvType) {
		mVal, mErr := GetMapValue(val)
		if mErr != nil {
			err = mErr
			return
		}

		item := &FieldValue{Name: key, Value: mVal}
		s.EqualFilter = append(s.EqualFilter, item)
		return
	}

	objVal, objErr := GetObjectValue(val)
	if objErr != nil {
		err = objErr
		return
	}

	item := &FieldValue{Name: key, Value: objVal}
	s.EqualFilter = append(s.EqualFilter, item)

	return
}

func (s *ObjectFilter) NotEqual(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := pu.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if om.IsSliceType(qvType) {
		err = fmt.Errorf("notEqual failed, illegal value type, type:%s", qv.Type().String())
		return
	}

	if qvType == om.TypeDateTimeValue {
		val = qv.Interface().(time.Time).Format(util.CSTLayout)
	}

	if om.IsBasicType(qvType) {
		item := &FieldValue{Name: key, Value: val}
		s.NotEqualFilter = append(s.NotEqualFilter, item)
		return
	}

	objVal, objErr := GetObjectValue(val)
	if objErr != nil {
		err = objErr
		return
	}

	item := &FieldValue{Name: key, Value: objVal}
	s.NotEqualFilter = append(s.NotEqualFilter, item)
	return nil
}

func (s *ObjectFilter) Below(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := pu.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !om.IsBasicType(qvType) {
		err = fmt.Errorf("below failed, illegal value type, type:%s", qv.Type().String())
		return
	}

	if qvType == om.TypeDateTimeValue {
		val = qv.Interface().(time.Time).Format(util.CSTLayout)
	}

	item := &FieldValue{Name: key, Value: val}
	s.BelowFilter = append(s.BelowFilter, item)

	return nil
}

func (s *ObjectFilter) Above(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := pu.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !om.IsBasicType(qvType) {
		err = fmt.Errorf("above failed, illegal value type, type:%s", qv.Type().String())
		return
	}

	if qvType == om.TypeDateTimeValue {
		val = qv.Interface().(time.Time).Format(util.CSTLayout)
	}

	item := &FieldValue{Name: key, Value: val}
	s.AboveFilter = append(s.AboveFilter, item)

	return nil
}

func (s *ObjectFilter) getSliceValue(sliceVal interface{}) (ret interface{}, err error) {
	sliceReVal := reflect.Indirect(reflect.ValueOf(sliceVal))
	sliceValType, sliceValErr := pu.GetTypeEnum(sliceReVal.Type())
	if sliceValErr != nil {
		err = sliceValErr
		return
	}

	if !om.IsSliceType(sliceValType) {
		err = fmt.Errorf("illegal value type, type:%s", sliceReVal.Type().String())
		return
	}

	if sliceReVal.Len() == 0 {
		return
	}

	svType := sliceReVal.Type().Elem()
	if svType.Kind() == reflect.Ptr {
		svType = svType.Elem()
	}

	subType, subErr := pu.GetTypeEnum(svType)
	if subErr != nil {
		err = subErr
		return
	}

	if om.IsStructType(subType) {
		ret, err = GetSliceObjectValue(sliceVal)
		return
	}

	retVal := []interface{}{}
	for idx := 0; idx < sliceReVal.Len(); idx++ {
		subV := reflect.Indirect(sliceReVal.Index(idx))
		if om.TypeDateTimeValue == subType {
			dtVal := subV.Interface().(time.Time).Format(util.CSTLayout)
			retVal = append(retVal, dtVal)

			continue
		}

		retVal = append(retVal, subV.Interface())
	}
	ret = retVal

	return
}

func (s *ObjectFilter) In(key string, val interface{}) (err error) {
	sliceVal, sliceErr := s.getSliceValue(val)
	if sliceErr != nil {
		err = sliceErr
		return
	}

	item := &FieldValue{Name: key, Value: sliceVal}
	s.InFilter = append(s.InFilter, item)

	return
}

func (s *ObjectFilter) NotIn(key string, val interface{}) (err error) {
	sliceVal, sliceErr := s.getSliceValue(val)
	if sliceErr != nil {
		err = sliceErr
		return
	}

	item := &FieldValue{Name: key, Value: sliceVal}
	s.NotInFilter = append(s.NotInFilter, item)

	return nil
}

func (s *ObjectFilter) Like(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	if qv.Kind() != reflect.String {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	item := &FieldValue{Name: key, Value: val}
	s.LikeFilter = append(s.LikeFilter, item)

	return nil
}

func (s *ObjectFilter) Page(filter *util.Pagination) {
	s.PageFilter = filter
}

func (s *ObjectFilter) Sort(sorter *util.SortFilter) {
	s.SortFilter = sorter
}

func (s *ObjectFilter) ValueMask(val interface{}) (err error) {
	if val == nil {
		err = fmt.Errorf("illegal mask value")
		return
	}

	objectMask, objectOK := val.(*ObjectValue)
	if !objectOK {
		objVal, objErr := GetObjectValue(val)
		if objErr != nil {
			err = objErr
			return
		}

		objectMask = objVal
	}

	s.MaskValue = objectMask
	return
}

func (s *ObjectFilter) GetFilterItem(key string) om.FilterItem {
	itemVal, itemErr := s.getFilterValue(key, s.EqualFilter)
	if itemErr != nil {
		return nil
	}
	if itemVal != nil {
		return &filterItem{oprCode: om.EqualOpr, value: pu.NewValue(itemVal.Get())}
	}

	itemVal, itemErr = s.getFilterValue(key, s.NotEqualFilter)
	if itemErr != nil {
		return nil
	}
	if itemVal != nil {
		return &filterItem{oprCode: om.NotEqualOpr, value: pu.NewValue(itemVal.Get())}
	}

	itemVal, itemErr = s.getFilterValue(key, s.BelowFilter)
	if itemErr != nil {
		return nil
	}
	if itemVal != nil {
		return &filterItem{oprCode: om.BelowOpr, value: pu.NewValue(itemVal.Get())}
	}

	itemVal, itemErr = s.getFilterValue(key, s.AboveFilter)
	if itemErr != nil {
		return nil
	}
	if itemVal != nil {
		return &filterItem{oprCode: om.AboveOpr, value: pu.NewValue(itemVal.Get())}
	}

	itemVal, itemErr = s.getFilterValue(key, s.InFilter)
	if itemErr != nil {
		return nil
	}
	if itemVal != nil {
		return &filterItem{oprCode: om.InOpr, value: pu.NewValue(itemVal.Get())}
	}

	itemVal, itemErr = s.getFilterValue(key, s.NotInFilter)
	if itemErr != nil {
		return nil
	}
	if itemVal != nil {
		return &filterItem{oprCode: om.NotInOpr, value: pu.NewValue(itemVal.Get())}
	}

	itemVal, itemErr = s.getFilterValue(key, s.LikeFilter)
	if itemErr != nil {
		return nil
	}
	if itemVal != nil {
		return &filterItem{oprCode: om.LikeOpr, value: pu.NewValue(itemVal.Get())}
	}

	return nil
}

func (s *ObjectFilter) getFilterValue(key string, items []*FieldValue) (ret *FieldValue, err error) {
	for _, val := range items {
		if key == val.Name {
			ret = val
			break
		}
	}

	if ret != nil {
		ret, err = ConvertItem(ret)
	}
	return
}

func (s *ObjectFilter) Pagination() (limit, offset int, paging bool) {
	paging = false
	if s.PageFilter == nil {
		return
	}

	paging = true
	limit = s.PageFilter.PageSize
	offset = s.PageFilter.PageSize * (s.PageFilter.PageNum - 1)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 100
	}

	return
}

func (s *ObjectFilter) Sorter() om.Sorter {
	if s.SortFilter == nil {
		return nil
	}

	return s.SortFilter
}

func (s *ObjectFilter) MaskModel() om.Model {
	maskObject := s.bindObject.Copy()
	if s.MaskValue != nil {
		for _, val := range s.MaskValue.Fields {
			maskObject.SetFieldValue(val.Name, val)
		}
	}

	return maskObject
}
