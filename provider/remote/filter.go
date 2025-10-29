package remote

import (
	"encoding/json"
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/utils"
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
	Name           string            `json:"name"`
	PkgPath        string            `json:"pkgPath"`
	EqualFilter    []*FieldValue     `json:"equal"`
	NotEqualFilter []*FieldValue     `json:"noEqual"`
	BelowFilter    []*FieldValue     `json:"below"`
	AboveFilter    []*FieldValue     `json:"above"`
	InFilter       []*FieldValue     `json:"in"`
	NotInFilter    []*FieldValue     `json:"notIn"`
	LikeFilter     []*FieldValue     `json:"like"`
	MaskValue      *ObjectValue      `json:"maskValue"`
	PageFilter     *utils.Pagination `json:"page"`
	SortFilter     *utils.SortFilter `json:"sort"`

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

func (s *ObjectFilter) Equal(key string, val any) (err *cd.Error) {
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
		if item != nil {
			s.EqualFilter = append(s.EqualFilter, item)
		}
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("equal failed, illegal value, key:%v, val:%v", key, val))
	}

	return
}

func (s *ObjectFilter) NotEqual(key string, val any) (err *cd.Error) {
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
		if item != nil {
			s.NotEqualFilter = append(s.NotEqualFilter, item)
		}
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("not equal failed, illegal value, key:%v, val:%v", key, val))
	}

	return
}

func (s *ObjectFilter) Below(key string, val any) (err *cd.Error) {
	switch val.(type) {
	case int8, int16, int32, int, int64,
		uint8, uint16, uint32, uint, uint64,
		float32, float64:
		item := &FieldValue{Name: key, Value: val}
		s.BelowFilter = append(s.BelowFilter, item)
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("below failed, illegal value, key:%v, val:%v", key, val))
	}
	return
}

func (s *ObjectFilter) Above(key string, val any) (err *cd.Error) {
	switch val.(type) {
	case int8, int16, int32, int, int64,
		uint8, uint16, uint32, uint, uint64,
		float32, float64:
		item := &FieldValue{Name: key, Value: val}
		s.AboveFilter = append(s.AboveFilter, item)
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("above failed, illegal value, key:%v, val:%v", key, val))
	}
	return
}

func (s *ObjectFilter) In(key string, val any) (err *cd.Error) {
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
		if item != nil {
			s.InFilter = append(s.InFilter, item)
		}
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("in failed, illegal value, key:%v, val:%v", key, val))
	}

	return
}

func (s *ObjectFilter) NotIn(key string, val any) (err *cd.Error) {
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
		if item != nil {
			s.NotInFilter = append(s.NotInFilter, item)
		}
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("not in failed, illegal value, key:%v, val:%v", key, val))
	}

	return
}

func (s *ObjectFilter) Like(key string, val any) (err *cd.Error) {
	switch val.(type) {
	case string:
		item := &FieldValue{Name: key, Value: val}
		s.LikeFilter = append(s.LikeFilter, item)
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("like failed, illegal value, key:%v, val:%v", key, val))
	}

	return
}

func (s *ObjectFilter) Pagination(pageNum, pageSize int) {
	s.PageFilter = &utils.Pagination{
		PageNum:  pageNum,
		PageSize: pageSize,
	}
}

func (s *ObjectFilter) Sort(fieldName string, ascFlag bool) {
	s.SortFilter = &utils.SortFilter{
		FieldName: fieldName,
		AscFlag:   ascFlag,
	}
}

func (s *ObjectFilter) ValueMask(val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.Unexpected, "illegal mask value")
		return
	}

	var objectValuePtr *ObjectValue
	switch v := val.(type) {
	case *ObjectValue:
		if v != nil {
			objectValuePtr = v
		}
	case ObjectValue:
		objectValuePtr = &v
	case json.RawMessage:
		objectValuePtr, err = DecodeObjectValue(v)
	default:
		err = cd.NewError(cd.Unexpected, "illegal mask value")
	}

	if err != nil {
		log.Errorf("ValueMask failed, err:%v", err.Error())
		return
	}

	if objectValuePtr == nil {
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

func (s *ObjectFilter) getFilterValue(key string, items []*FieldValue) (ret *FieldValue, err *cd.Error) {
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

func (s *ObjectFilter) Paginationer() model.Paginationer {
	if s.PageFilter == nil {
		return nil
	}

	return s.PageFilter
}

func (s *ObjectFilter) Sorter() model.Sorter {
	if s.SortFilter == nil {
		return nil
	}

	return s.SortFilter
}

func (s *ObjectFilter) MaskModel() model.Model {
	maskObject := s.bindObject
	if s.MaskValue != nil {
		for _, val := range s.MaskValue.Fields {
			maskObject.SetFieldValue(val.Name, val.GetValue().Get())
		}
	}

	return maskObject.Copy(model.OriginView)
}
