package remote

import (
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/util"
)

// Item Item
type Item struct {
	Index int    `json:"index"`
	Name  string `json:"name"`

	Tag   TagImpl  `json:"tag"`
	Type  TypeImpl `json:"type"`
	value ValueImpl
}

// GetIndex GetIndex
func (s *Item) GetIndex() (ret int) {
	return s.Index
}

// GetName GetName
func (s *Item) GetName() string {
	return s.Name
}

// GetType GetType
func (s *Item) GetType() (ret model.Type) {
	ret = &s.Type
	return
}

// GetTag GetTag
func (s *Item) GetTag() (ret model.Tag) {
	ret = &s.Tag
	return
}

// GetValue GetValue
func (s *Item) GetValue() (ret model.Value) {
	ret = &s.value
	return
}

// IsPrimary IsPrimary
func (s *Item) IsPrimary() bool {
	return s.Tag.IsPrimaryKey()
}

// IsAssigned IsAssigned
func (s *Item) IsAssigned() (ret bool) {
	ret = false
	if s.value.IsNil() {
		return
	}

	currentVal := s.value.Get()
	originVal := s.Type.Interface()
	if util.IsBasicType(s.Type.GetValue()) {
		sameVal, sameErr := util.IsSameVal(originVal, currentVal)
		if sameErr != nil {
			log.Errorf("compare value failed, err:%s", sameErr.Error())
			ret = false
			return
		}

		// 值不相等，则可以认为有赋值过
		ret = !sameVal
		return
	}

	if util.IsStructType(s.Type.GetValue()) {
		if s.Type.IsPtrType() {
			ret = true
			return
		}

		curObj, curOK := currentVal.Interface().(ObjectValue)
		if !curOK {
			log.Errorf("illegal item value. val:%v", currentVal.Interface())
			ret = false
		} else {
			ret = curObj.IsAssigned()
		}

		return
	}

	if util.IsSliceType(s.Type.GetValue()) {
		if currentVal.Len() > 0 {
			ret = true
			return
		}
	}

	log.Errorf("illegal item type value, type:%s, value:%d", s.Type.GetName(), s.Type.GetValue())
	return
}

// SetValue SetValue
func (s *Item) SetValue(val reflect.Value) (err error) {
	toVal := s.Type.Interface()
	toVal, err = helper.AssignValue(val, toVal)
	if err != nil {
		log.Errorf("assign value failed, name:%s, err:%s", s.Name, err.Error())
		return
	}

	err = s.value.Set(toVal)
	if err != nil {
		log.Errorf("set value failed, name:%s, err:%s", s.Name, err.Error())
	}

	return
}

// UpdateValue UpdateValue
func (s *Item) UpdateValue(val reflect.Value) (err error) {
	err = s.value.Update(val)
	if err != nil {
		log.Errorf("update value failed, name:%s, err:%s", s.Name, err.Error())
	}

	return
}

// Copy Copy
func (s *Item) Copy() (ret model.Field) {
	return &Item{Index: s.Index, Name: s.Name, Tag: s.Tag, Type: s.Type, value: s.value}
}

// Interface interface value
func (s *Item) Interface() (ret *ItemValue) {
	if s.Type.IsPtrType() {
		ret = &ItemValue{Name: s.Name, Value: nil}
		return
	}

	ret = &ItemValue{Name: s.Name, Value: s.Type.Interface().Interface()}
	return
}
