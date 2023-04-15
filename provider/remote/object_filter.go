package remote

import (
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
	ou "github.com/muidea/magicOrm/util"
)

// ObjectFilter value filter
type ObjectFilter struct {
	Name           string           `json:"name"`
	PkgPath        string           `json:"pkgPath"`
	EqualFilter    []*ItemValue     `json:"equal"`
	NotEqualFilter []*ItemValue     `json:"noEqual"`
	BelowFilter    []*ItemValue     `json:"below"`
	AboveFilter    []*ItemValue     `json:"above"`
	InFilter       []*ItemValue     `json:"in"`
	NotInFilter    []*ItemValue     `json:"notIn"`
	LikeFilter     []*ItemValue     `json:"like"`
	MaskValue      *ObjectValue     `json:"maskValue"`
	PageFilter     *util.Pagination `json:"page"`
	SortFilter     *util.SortFilter `json:"sort"`
}

// NewFilter new query filter
func NewFilter() *ObjectFilter {
	return &ObjectFilter{
		EqualFilter:    []*ItemValue{},
		NotEqualFilter: []*ItemValue{},
		BelowFilter:    []*ItemValue{},
		AboveFilter:    []*ItemValue{},
		InFilter:       []*ItemValue{},
		NotInFilter:    []*ItemValue{},
		LikeFilter:     []*ItemValue{},
	}
}

func (s *ObjectFilter) FromHttpRequest(req *http.Request) {
	filter := util.NewFilter()
	filter.Decode(req)

	s.FromContentFilter(filter)
}

func (s *ObjectFilter) FromContentFilter(filter *util.ContentFilter) {
	if filter == nil {
		return
	}

	if filter.ParamItems != nil {
		for k, _ := range filter.ParamItems.Items {
			val := filter.ParamItems.GetEqual(k)
			if val != nil {
				s.Equal(k, val)
				continue
			}
			val = filter.ParamItems.GetNotEqual(k)
			if val != nil {
				s.NotEqual(k, val)
				continue
			}
			val = filter.ParamItems.GetBelow(k)
			if val != nil {
				s.Below(k, val)
				continue
			}
			val = filter.ParamItems.GetAbove(k)
			if val != nil {
				s.Above(k, val)
				continue
			}
			val = filter.ParamItems.GetIn(k)
			if val != nil {
				s.In(k, val)
				continue
			}
			val = filter.ParamItems.GetNotIn(k)
			if val != nil {
				s.NotIn(k, val)
				continue
			}
			val = filter.ParamItems.GetLike(k)
			if val != nil {
				s.Like(k, val)
				continue
			}
		}
	}

	if filter.Pagination != nil {
		s.PageFilter = filter.Pagination
		return
	}

	s.PageFilter = util.DefaultPagination()
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

func (s *ObjectFilter) Pagination(pageFilter *util.Pagination) {
	if pageFilter == nil {
		return
	}

	s.PageFilter = pageFilter
}

// Equal assign equal filter value
func (s *ObjectFilter) Equal(key string, val interface{}) (err error) {
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
		val = qv.Interface().(time.Time).Format(time.RFC3339)
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
func (s *ObjectFilter) NotEqual(key string, val interface{}) (err error) {
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
func (s *ObjectFilter) Below(key string, val interface{}) (err error) {
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
func (s *ObjectFilter) Above(key string, val interface{}) (err error) {
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

func (s *ObjectFilter) getSliceValue(sliceVal interface{}) (ret interface{}, err error) {
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
func (s *ObjectFilter) In(key string, val interface{}) (err error) {
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
func (s *ObjectFilter) NotIn(key string, val interface{}) (err error) {
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
func (s *ObjectFilter) Like(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	if qv.Kind() != reflect.String {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	item := &ItemValue{Name: key, Value: val}
	s.LikeFilter = append(s.LikeFilter, item)

	return nil
}

func (s *ObjectFilter) ValueMask(val interface{}) (err error) {
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

func (s *ObjectFilter) Page(filter *util.Pagination) {
	s.PageFilter = filter
}

func (s *ObjectFilter) Sort(sorter *util.SortFilter) {
	s.SortFilter = sorter
}
