package remote

import (
	"log"
	"reflect"

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
	if currentVal.Kind() == reflect.Interface {
		currentVal = currentVal.Elem()
	}
	// 非空指针，则表示已经赋值
	if s.Type.IsPtrType() {
		currentVal = reflect.Indirect(currentVal)
	}
	if util.IsNil(currentVal) {
		return
	}
	if util.IsBasicType(s.Type.GetValue()) {
		originVal := reflect.New(currentVal.Type()).Elem()
		sameVal, sameErr := util.IsSameVal(originVal, currentVal)
		if sameErr != nil {
			log.Printf("compare value failed, err:%s", sameErr.Error())
			ret = false
			return
		}
		// 值不相等，则可以认为有赋值过
		ret = !sameVal
	} else if util.IsStructType(s.Type.GetValue()) {
		curObj, curOK := currentVal.Interface().(ObjectValue)
		if !curOK {
			log.Fatalf("illegal item value. val:%v", currentVal.Interface())
			ret = false
		} else {
			ret = curObj.IsAssigned()
		}
	} else if util.IsSliceType(s.Type.GetValue()) {
		if currentVal.Len() > 0 {
			ret = true
			return
		}
	} else {
		log.Fatalf("illegal item type value, type:%s, value:%d", s.Type.GetName(), s.Type.GetValue())
	}

	return
}

// SetValue SetValue
func (s *Item) SetValue(val reflect.Value) (err error) {
	err = s.value.Set(val)
	return
}

// UpdateValue UpdateValue
func (s *Item) UpdateValue(val reflect.Value) (err error) {
	val = reflect.Indirect(val)
	depend := s.Type.Depend()
	if depend == nil || util.IsBasicType(depend.GetValue()) {
		fieldVal := reflect.Indirect(s.Type.Interface())
		valErr := helper.ConvertValue(val, &fieldVal)
		if valErr != nil {
			err = valErr
			log.Printf("helper.ConvertValue failed, err:%s", err.Error())
			return
		}

		if s.Type.IsPtrType() {
			fieldVal = fieldVal.Addr()
		}

		err = s.value.Update(fieldVal)
		return
	}

	err = s.value.Update(val)
	return
}

// Copy Copy
func (s *Item) Copy() (ret model.Field) {
	return &Item{Index: s.Index, Name: s.Name, Tag: s.Tag, Type: s.Type}
}

// Interface interface value
func (s *Item) Interface() (ret *ItemValue) {
	if s.value.IsNil() {
		ret = &ItemValue{Name: s.Name, Value: false}
	}
	ret = &ItemValue{Name: s.Name, Value: s.Type.Interface().Interface()}

	return
}
