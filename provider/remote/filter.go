package remote

import (
	"encoding/json"
	"fmt"

	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/model"
)

type filterItem struct {
	oprCode model.OprCode
	value   *ValueImpl
}

func (s *filterItem) OprCode() model.OprCode {
	return s.oprCode
}

func (s *filterItem) OprValue() model.Value {
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
	switch val.(type) {
	case bool,
		int8, int16, int32, int, int64,
		uint8, uint16, uint32, uint, uint64,
		float32, float64,
		string,
		map[string]any,
		*ObjectValue:
		item := &FieldValue{Name: key, Value: val}
		item, err = ConvertItem(item)
		if err != nil {
			return
		}
		s.EqualFilter = append(s.EqualFilter, item)
	default:
		err := fmt.Errorf("equal failed, illegal value, val:%v", val)
		panic(err.Error())
	}

	return
}

func (s *ObjectFilter) NotEqual(key string, val interface{}) (err error) {
	switch val.(type) {
	case bool,
		int8, int16, int32, int, int64,
		uint8, uint16, uint32, uint, uint64,
		float32, float64,
		string,
		map[string]any,
		*ObjectValue:
		item := &FieldValue{Name: key, Value: val}
		item, err = ConvertItem(item)
		if err != nil {
			return
		}
		s.NotEqualFilter = append(s.NotEqualFilter, item)
	default:
		err := fmt.Errorf("not equal failed, illegal value, val:%v", val)
		panic(err.Error())
	}

	return
}

func (s *ObjectFilter) Below(key string, val interface{}) (err error) {
	switch val.(type) {
	case int8, int16, int32, int, int64,
		uint8, uint16, uint32, uint, uint64,
		float32, float64:
		item := &FieldValue{Name: key, Value: val}
		s.BelowFilter = append(s.BelowFilter, item)
	default:
		err := fmt.Errorf("below failed, illegal value, val:%v", val)
		panic(err.Error())
	}
	return
}

func (s *ObjectFilter) Above(key string, val interface{}) (err error) {
	switch val.(type) {
	case int8, int16, int32, int, int64,
		uint8, uint16, uint32, uint, uint64,
		float32, float64:
		item := &FieldValue{Name: key, Value: val}
		s.AboveFilter = append(s.AboveFilter, item)
	default:
		err := fmt.Errorf("below failed, illegal value, val:%v", val)
		panic(err.Error())
	}
	return
}

func (s *ObjectFilter) In(key string, val interface{}) (err error) {
	switch val.(type) {
	case []bool,
		[]int8, []int16, []int32, []int, []int64,
		[]uint8, []uint16, []uint32, []uint, []uint64,
		[]float32, []float64,
		[]string,
		[]any,
		map[string]any,
		*SliceObjectValue:
		item := &FieldValue{Name: key, Value: val}
		item, err = ConvertItem(item)
		if err != nil {
			return
		}
		s.InFilter = append(s.InFilter, item)
	default:
		err := fmt.Errorf("in failed, illegal value, val:%v", val)
		panic(err.Error())
	}

	return
}

func (s *ObjectFilter) NotIn(key string, val interface{}) (err error) {
	switch val.(type) {
	case []bool,
		[]int8, []int16, []int32, []int, []int64,
		[]uint8, []uint16, []uint32, []uint, []uint64,
		[]float32, []float64,
		[]string,
		[]any,
		map[string]any,
		*SliceObjectValue:
		item := &FieldValue{Name: key, Value: val}
		item, err = ConvertItem(item)
		if err != nil {
			return
		}
		s.NotInFilter = append(s.NotInFilter, item)
	default:
		err := fmt.Errorf("not in failed, illegal value, val:%v", val)
		panic(err.Error())
	}

	return
}

func (s *ObjectFilter) Like(key string, val interface{}) (err error) {
	switch val.(type) {
	case string:
		item := &FieldValue{Name: key, Value: val}
		s.LikeFilter = append(s.LikeFilter, item)
	default:
		err := fmt.Errorf("like failed, illegal value, val:%v", val)
		panic(err.Error())
	}

	return
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

	var objectValuePtr *ObjectValue
	switch val.(type) {
	case *ObjectValue:
		objectMask, objectOK := val.(*ObjectValue)
		if objectOK {
			objectValuePtr = objectMask
		}
	case json.RawMessage:
		byteVal, byteOK := val.(json.RawMessage)
		if byteOK {
			objectValuePtr, err = DecodeObjectValue(byteVal)
		}
	}
	if objectValuePtr == nil {
		err = fmt.Errorf("illegal mask value")
		log.Errorf("ValueMask failed, err:%v", err.Error())
		return
	}

	if s.bindObject.GetPkgKey() != objectValuePtr.GetPkgKey() {
		err = fmt.Errorf("mismatch mask value, bindPkgKey:%v, maskPkgKey:%v", s.bindObject.GetPkgKey(), objectValuePtr.GetPkgKey())
		log.Errorf("ValueMask failed, err:%v", err.Error())
		return
	}

	s.MaskValue = objectValuePtr
	return
}

func (s *ObjectFilter) GetFilterItem(key string) model.FilterItem {
	itemVal, itemErr := s.getFilterValue(key, s.EqualFilter)
	if itemErr != nil {
		return nil
	}
	if itemVal != nil {
		return &filterItem{oprCode: model.EqualOpr, value: NewValue(itemVal.Get())}
	}

	itemVal, itemErr = s.getFilterValue(key, s.NotEqualFilter)
	if itemErr != nil {
		return nil
	}
	if itemVal != nil {
		return &filterItem{oprCode: model.NotEqualOpr, value: NewValue(itemVal.Get())}
	}

	itemVal, itemErr = s.getFilterValue(key, s.BelowFilter)
	if itemErr != nil {
		return nil
	}
	if itemVal != nil {
		return &filterItem{oprCode: model.BelowOpr, value: NewValue(itemVal.Get())}
	}

	itemVal, itemErr = s.getFilterValue(key, s.AboveFilter)
	if itemErr != nil {
		return nil
	}
	if itemVal != nil {
		return &filterItem{oprCode: model.AboveOpr, value: NewValue(itemVal.Get())}
	}

	itemVal, itemErr = s.getFilterValue(key, s.InFilter)
	if itemErr != nil {
		return nil
	}
	if itemVal != nil {
		return &filterItem{oprCode: model.InOpr, value: NewValue(itemVal.Get())}
	}

	itemVal, itemErr = s.getFilterValue(key, s.NotInFilter)
	if itemErr != nil {
		return nil
	}
	if itemVal != nil {
		return &filterItem{oprCode: model.NotInOpr, value: NewValue(itemVal.Get())}
	}

	itemVal, itemErr = s.getFilterValue(key, s.LikeFilter)
	if itemErr != nil {
		return nil
	}
	if itemVal != nil {
		return &filterItem{oprCode: model.LikeOpr, value: NewValue(itemVal.Get())}
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

func (s *ObjectFilter) Sorter() model.Sorter {
	if s.SortFilter == nil {
		return nil
	}

	return s.SortFilter
}

func (s *ObjectFilter) MaskModel() model.Model {
	reset := s.MaskValue != nil
	maskObject := s.bindObject.Copy(reset)
	if reset {
		for _, val := range s.MaskValue.Fields {
			err := maskObject.SetFieldValue(val.Name, val.GetValue())
			if err != nil {
				log.Errorf("MaskModel failed, maskObject.SetFieldValue error:%s", err.Error())
			}
		}
	}

	return maskObject
}
