package remote

import (
	"fmt"
	"github.com/muidea/magicOrm/provider/helper"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// Item Item
type Item struct {
	Index int    `json:"index"`
	Name  string `json:"name"`

	Tag   *TagImpl  `json:"tag"`
	Type  *TypeImpl `json:"type"`
	value *ValueImpl
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
	if s.Type != nil {
		ret = s.Type
	}
	return
}

// GetTag GetTag
func (s *Item) GetTag() (ret model.Tag) {
	if s.Tag != nil {
		ret = s.Tag
	}

	return
}

// GetValue GetValue
func (s *Item) GetValue() (ret model.Value) {
	if s.value != nil {
		ret = s.value
	}

	return
}

// IsPrimary IsPrimary
func (s *Item) IsPrimary() bool {
	return s.Tag.IsPrimaryKey()
}

// IsAssigned IsAssigned
func (s *Item) IsAssigned() (ret bool) {
	ret = false
	if s.value == nil {
		return
	}

	if s.value.IsNil() {
		return
	}

	currentVal := s.value.Get()
	originVal := s.Type.Interface()
	if util.IsBasicType(s.Type.GetValue()) {
		sameVal, sameErr := util.IsSameVal(originVal, currentVal)
		if sameErr != nil {
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

		curObj, curOK := currentVal.Interface().(*ObjectValue)
		if !curOK {
			log.Errorf("illegal item value. val:%v", currentVal.Interface())
			ret = false
		} else {
			ret = curObj.IsAssigned()
		}

		return
	}

	if util.IsSliceType(s.Type.GetValue()) {
		if s.Type.IsPtrType() {
			ret = true
			return
		}

		curObj, curOK := currentVal.Interface().(*SliceObjectValue)
		if !curOK {
			log.Errorf("illegal slice item value. val:%v", currentVal.Interface())
			ret = false
		} else {
			ret = len(curObj.Values) > 0
		}

		return
	}

	return
}

// SetValue SetValue
func (s *Item) SetValue(val reflect.Value) (err error) {
	if s.value == nil {
		vVal, vErr := newValue(val)
		if vErr != nil {
			err = vErr
			return
		}

		s.value = vVal
		return
	}

	err = s.value.Set(val)
	if err != nil {
		log.Errorf("set item value failed, name:%s, err:%s", s.Name, err.Error())
	}

	return
}

// UpdateValue UpdateValue
func (s *Item) UpdateValue(val reflect.Value) (err error) {
	if util.IsNil(val) {
		err = fmt.Errorf("invalid update value")
		return
	}

	toVal := s.Type.Interface()

	dependType := s.Type.Depend()
	if util.IsBasicType(s.GetType().GetValue()) || util.IsBasicType(dependType.GetValue()) {
		toVal, err = helper.AssignValue(val, toVal)
		if err != nil {
			log.Errorf("assign value failed, name:%s, from type:%s, to type:%s, err:%s", s.Name, val.Type().String(), s.Type.GetName(), err.Error())
			return
		}

		err = s.value.Update(toVal)
		return
	}

	err = s.value.Update(val)
	if err != nil {
		log.Errorf("update item value failed, name:%s, err:%s", s.Name, err.Error())
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

func newItem(idx int, name string, iTag *TagImpl, iType *TypeImpl) *Item {
	return &Item{Index: idx, Name: name, Tag: iTag, Type: iType}
}

func compareItem(l, r *Item) bool {
	if l.Index != r.Index {
		return false
	}
	if l.Name != r.Name {
		return false
	}

	if !compareType(l.Type, r.Type) {
		return false
	}
	if !compareTag(l.Tag, r.Tag) {
		return false
	}

	return true
}
