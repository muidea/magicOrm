package remote

import (
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
	ou "github.com/muidea/magicOrm/util"
)

// SortItem 排序项
type SortItem struct {
	// true:升序,false:降序
	Asc bool `json:"asc"`
	// 排序字段
	Name string `json:"name"`
}

// Pagination 页面过滤器
type Pagination struct {
	// 单页条目数
	Size int `json:"size"`
	// 页码
	Num int `json:"num"`
}

// QueryFilter value filter
type QueryFilter struct {
	EqualFilter    []*ItemValue `json:"equal"`
	NotEqualFilter []*ItemValue `json:"noEqual"`
	BelowFilter    []*ItemValue `json:"below"`
	AboveFilter    []*ItemValue `json:"above"`
	InFilter       []*ItemValue `json:"in"`
	NotInFilter    []*ItemValue `json:"notIn"`
	LikeFilter     []*ItemValue `json:"like"`
	PageFilter     *Pagination  `json:"page"`
	SortFilter     *SortItem    `json:"sort"`
	MaskValue      *ObjectValue `json:"maskValue"`
}

// NewFilter new query filter
func NewFilter() *QueryFilter {
	return &QueryFilter{
		EqualFilter:    []*ItemValue{},
		NotEqualFilter: []*ItemValue{},
		BelowFilter:    []*ItemValue{},
		AboveFilter:    []*ItemValue{},
		InFilter:       []*ItemValue{},
		NotInFilter:    []*ItemValue{},
		LikeFilter:     []*ItemValue{},
	}
}

func (s *QueryFilter) GetString(key string) (ret string, ok bool) {
	for _, item := range s.EqualFilter {
		if item.Name == key {
			ret, ok = (item.Value).(string)
			return
		}
	}

	return
}

func (s *QueryFilter) GetInt(key string) (ret int, ok bool) {
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

// Equal assign equal filter value
func (s *QueryFilter) Equal(key string, val interface{}) (err error) {
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

	if qvType == ou.TypeDateTimeField {
		val = qv.Interface().(time.Time).Format(util.CSTLayout)
	}

	if ou.IsBasicType(qvType) {
		item := &ItemValue{Name: key, Value: val}
		s.EqualFilter = append(s.EqualFilter, item)
		return
	}

	if ou.IsMapType(qvType) {
		mVal, mErr := GetMapValue(val)
		if mErr != nil {
			err = mErr
			return
		}

		item := &ItemValue{Name: key, Value: mVal}
		s.EqualFilter = append(s.EqualFilter, item)
		return
	}

	objVal, objErr := GetObjectValue(val)
	if objErr != nil {
		err = objErr
		return
	}

	item := &ItemValue{Name: key, Value: objVal}
	s.EqualFilter = append(s.EqualFilter, item)

	return
}

// NotEqual  assign no equal filter value
func (s *QueryFilter) NotEqual(key string, val interface{}) (err error) {
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

	if qvType == ou.TypeDateTimeField {
		val = qv.Interface().(time.Time).Format(util.CSTLayout)
	}

	if ou.IsBasicType(qvType) {
		item := &ItemValue{Name: key, Value: val}
		s.NotEqualFilter = append(s.NotEqualFilter, item)
		return
	}

	objVal, objErr := GetObjectValue(val)
	if objErr != nil {
		err = objErr
		return
	}

	item := &ItemValue{Name: key, Value: objVal}
	s.NotEqualFilter = append(s.NotEqualFilter, item)
	return nil
}

// Below assign below filter value
func (s *QueryFilter) Below(key string, val interface{}) (err error) {
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

	if qvType == ou.TypeDateTimeField {
		val = qv.Interface().(time.Time).Format(util.CSTLayout)
	}

	item := &ItemValue{Name: key, Value: val}
	s.BelowFilter = append(s.BelowFilter, item)

	return nil
}

// Above assign above filter value
func (s *QueryFilter) Above(key string, val interface{}) (err error) {
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

	if qvType == ou.TypeDateTimeField {
		val = qv.Interface().(time.Time).Format("2006-01-02 15:04:05")
	}

	item := &ItemValue{Name: key, Value: val}
	s.AboveFilter = append(s.AboveFilter, item)

	return nil
}

func (s *QueryFilter) getSliceValue(sliceVal interface{}) (ret interface{}, err error) {
	sliceReVal := reflect.Indirect(reflect.ValueOf(sliceVal))
	sliceValType, sliceValErr := ou.GetTypeEnum(sliceReVal.Type())
	if sliceValErr != nil {
		err = sliceValErr
		return
	}

	if !ou.IsSliceType(sliceValType) {
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

	subType, subErr := ou.GetTypeEnum(svType)
	if subErr != nil {
		err = subErr
		return
	}

	if ou.IsStructType(subType) {
		ret, err = GetSliceObjectValue(sliceVal)
		return
	}

	retVal := []interface{}{}
	for idx := 0; idx < sliceReVal.Len(); idx++ {
		subV := reflect.Indirect(sliceReVal.Index(idx))
		if ou.TypeDateTimeField == subType {
			dtVal := subV.Interface().(time.Time).Format("2006-01-02 15:04:05")
			retVal = append(retVal, dtVal)

			continue
		}

		retVal = append(retVal, subV.Interface())
	}
	ret = retVal

	return
}

// In assign in filter value
func (s *QueryFilter) In(key string, val interface{}) (err error) {
	sliceVal, sliceErr := s.getSliceValue(val)
	if sliceErr != nil {
		err = sliceErr
		return
	}

	item := &ItemValue{Name: key, Value: sliceVal}
	s.InFilter = append(s.InFilter, item)

	return
}

// NotIn assign notIn filter value
func (s *QueryFilter) NotIn(key string, val interface{}) (err error) {
	sliceVal, sliceErr := s.getSliceValue(val)
	if sliceErr != nil {
		err = sliceErr
		return
	}

	item := &ItemValue{Name: key, Value: sliceVal}
	s.NotInFilter = append(s.NotInFilter, item)

	return nil
}

// Like assign like filter value
func (s *QueryFilter) Like(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	if qv.Kind() != reflect.String {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	item := &ItemValue{Name: key, Value: val}
	s.LikeFilter = append(s.LikeFilter, item)

	return nil
}

func (s *QueryFilter) Page(filter *util.Pagination) {
	if filter == nil {
		return
	}

	s.PageFilter = &Pagination{
		Num:  filter.PageNum,
		Size: filter.PageSize,
	}
}

func (s *QueryFilter) Sort(sorter *util.SortFilter) {
	if sorter == nil {
		return
	}

	s.SortFilter = &SortItem{
		Name: sorter.FieldName,
		Asc:  sorter.AscFlag,
	}
}

// ValueMask assign mask value
func (s *QueryFilter) ValueMask(val interface{}) (err error) {
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

	objVal, objErr := GetObjectValue(val)
	if objErr != nil {
		err = objErr
		return
	}

	s.MaskValue = objVal
	return
}

// ObjectValueFilter object value filter
type ObjectValueFilter struct {
	TypeName    string       `json:"typeName"`
	PkgPath     string       `json:"pkgPath"`
	ValueFilter *QueryFilter `json:"valueFilter"`
}

func (s *ObjectValueFilter) GetName() string {
	return s.TypeName
}

func (s *ObjectValueFilter) GetPkgPath() string {
	return s.PkgPath
}
