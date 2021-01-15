package local

import (
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

type typeImpl struct {
	typeImpl reflect.Type
}

// newType newType
func newType(val reflect.Type) (ret *typeImpl, err error) {
	rawType := val
	if rawType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
	}
	_, err = util.GetTypeEnum(rawType)
	if err != nil {
		return
	}

	ret = &typeImpl{typeImpl: val}
	return
}

func (s *typeImpl) GetName() string {
	tType := s.getType()
	if tType.Kind() == reflect.Slice {
		tType = tType.Elem()
	}
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}

	return tType.String()
}

func (s *typeImpl) GetValue() (ret int) {
	tType := s.getType()
	ret, _ = util.GetTypeEnum(tType)
	return
}

func (s *typeImpl) GetPkgPath() string {
	tType := s.getType()
	if tType.Kind() == reflect.Slice {
		tType = tType.Elem()
	}
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}

	return tType.PkgPath()
}

func (s *typeImpl) IsPtrType() bool {
	return s.typeImpl.Kind() == reflect.Ptr
}

func (s *typeImpl) Interface(val interface{}) (ret model.Value, err error) {
	tType := s.getType()
	tVal := reflect.New(tType).Elem()
	if val != nil {
		rVal := reflect.ValueOf(val)
		if rVal.Kind() == reflect.Interface {
			rVal = rVal.Elem()
		}

		assignFlag := false
		rVal = reflect.Indirect(rVal)
		rType := rVal.Type()
		if util.IsBool(tType) {
			if util.IsBool(rType) {
				tVal.SetBool(rVal.Bool())
				assignFlag = true
			}
			if util.IsInteger(rType) {
				tVal.SetBool(rVal.Int() > 0)
				assignFlag = true
			}
			if util.IsUInteger(rType) {
				tVal.SetBool(rVal.Uint() > 0)
				assignFlag = true
			}
		}
		if util.IsInteger(tType) && util.IsInteger(rType) {
			tVal.SetInt(rVal.Int())
			assignFlag = true
		}
		if util.IsUInteger(tType) && util.IsUInteger(rType) {
			tVal.SetUint(rVal.Uint())
			assignFlag = true
		}
		if util.IsFloat(tType) && util.IsFloat(rType) {
			tVal.SetFloat(rVal.Float())
			assignFlag = true
		}
		if util.IsString(tType) && util.IsString(rType) {
			tVal.SetString(rVal.String())
			assignFlag = true
		}
		if util.IsDateTime(tType) {
			if util.IsDateTime(rType) {
				tVal.Set(rVal)
				assignFlag = true
			}
			if util.IsString(rType) {
				dtVal, dtErr := time.Parse("2006-01-02 15:04:05", rVal.String())
				if dtErr == nil {
					tVal.Set(reflect.ValueOf(dtVal))
					assignFlag = true
				}
			}
		}
		if util.IsSlice(tType) {
			if util.IsString(rType) {
				sVal, sErr := _helper.Decode(rVal.String(), s)
				if sErr == nil {
					tVal.Set(sVal.Get().(reflect.Value))
					assignFlag = true
				}
			}
		}

		if tType.String() == rType.String() {
			tVal.Set(rVal)
			assignFlag = true
		}
		if !assignFlag {
			err = fmt.Errorf("illegal initialize value, current type:%s,expect type:%s", tType.String(), rType.String())
			return
		}
	}

	if s.IsPtrType() {
		tVal = tVal.Addr()
	}
	ret = newValue(tVal)
	return
}

func (s *typeImpl) Elem() model.Type {
	tType := s.getType()
	if tType.Kind() == reflect.Slice {
		return &typeImpl{typeImpl: tType.Elem()}
	}

	return &typeImpl{typeImpl: s.typeImpl}
}

func (s *typeImpl) IsBasic() bool {
	elemType := s.Elem()

	return util.IsBasicType(elemType.GetValue())
}

func (s *typeImpl) getType() reflect.Type {
	if s.typeImpl.Kind() == reflect.Ptr {
		return s.typeImpl.Elem()
	}

	return s.typeImpl
}

func (s *typeImpl) copy() (ret *typeImpl) {
	ret = &typeImpl{
		typeImpl: s.typeImpl,
	}

	return
}

func (s *typeImpl) dump() string {
	val := s.GetValue()
	return fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", val, s.GetName(), s.GetPkgPath(), s.IsPtrType())
}
