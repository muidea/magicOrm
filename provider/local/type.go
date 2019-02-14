package local

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

type typeImpl struct {
	typeImpl reflect.Type
}

func (s *typeImpl) GetName() string {
	if s.typeImpl.Kind() == reflect.Ptr {
		return s.typeImpl.Elem().Name()
	}
	return s.typeImpl.Name()
}

func (s *typeImpl) GetValue() (ret int) {
	if s.typeImpl.Kind() == reflect.Ptr {
		ret, _ = util.GetTypeValueEnum(s.typeImpl.Elem())
		return
	}

	ret, _ = util.GetTypeValueEnum(s.typeImpl)
	return
}

func (s *typeImpl) GetPkgPath() string {
	if s.typeImpl.Kind() == reflect.Ptr {
		return s.typeImpl.Elem().PkgPath()
	}
	return s.typeImpl.PkgPath()
}

func (s *typeImpl) GetType() reflect.Type {
	if s.typeImpl.Kind() == reflect.Ptr {
		return s.typeImpl.Elem()
	}

	return s.typeImpl
}

func (s *typeImpl) GetDepend() (ret reflect.Type) {
	rawVal := s.typeImpl
	if rawVal.Kind() == reflect.Ptr {
		rawVal = rawVal.Elem()
	}
	if rawVal.Kind() == reflect.Struct {
		if rawVal.String() != "time.Time" {
			ret = s.typeImpl
			return
		}
	}

	if rawVal.Kind() == reflect.Slice {
		sliceVal := rawVal.Elem()
		rawVal = sliceVal
		if rawVal.Kind() == reflect.Ptr {
			rawVal = rawVal.Elem()
		}
		if rawVal.Kind() == reflect.Struct {
			if rawVal.String() != "time.Time" {
				ret = sliceVal
				return
			}
		}
	}

	return
}

func (s *typeImpl) IsPtrType() bool {
	return s.typeImpl.Kind() == reflect.Ptr
}

func (s *typeImpl) String() string {
	return fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", s.GetValue(), s.GetName(), s.GetPkgPath(), s.IsPtrType())
}

func (s *typeImpl) Copy() model.FieldType {
	return &typeImpl{
		typeImpl: s.typeImpl,
	}
}

// NewFieldType NewFieldType
func NewFieldType(val reflect.Type) (ret model.FieldType, err error) {
	rawVal := val
	if rawVal.Kind() == reflect.Ptr {
		rawVal = rawVal.Elem()
	}
	if rawVal.Kind() != reflect.Slice {
		_, tErr := util.GetTypeValueEnum(rawVal)
		if tErr != nil {
			err = tErr
			return
		}
		ret = &typeImpl{typeImpl: val}
		return
	}

	// check slice elemnt type
	sliceVal := rawVal.Elem()
	if sliceVal.Kind() == reflect.Ptr {
		sliceVal = sliceVal.Elem()
	}
	if sliceVal.Kind() == reflect.Slice {
		err = fmt.Errorf("illegal slice element type")
		return
	}

	_, tErr := util.GetTypeValueEnum(sliceVal)
	if tErr != nil {
		err = tErr
		return
	}
	ret = &typeImpl{typeImpl: val}
	return
}
