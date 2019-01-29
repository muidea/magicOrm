package local

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

type typeStruct struct {
	typeValue   int
	typeName    string
	typePkgPath string
	typeIsPtr   bool
	typeDepend  reflect.Type
}

func (s *typeStruct) Name() string {
	return s.typeName
}

func (s *typeStruct) Value() int {
	return s.typeValue
}

func (s *typeStruct) IsPtr() bool {
	return s.typeIsPtr
}

func (s *typeStruct) PkgPath() string {
	return s.typePkgPath
}

func (s *typeStruct) String() string {
	ret := fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", s.typeValue, s.typeName, s.typePkgPath, s.typeIsPtr)
	if s.typeDepend != nil {
		ret = fmt.Sprintf("%s,depend:[%s]", ret, s.typeDepend)
	}

	return ret
}

func (s *typeStruct) Depend() reflect.Type {
	return s.typeDepend
}

func (s *typeStruct) Copy() model.FieldType {
	return &typeSlice{
		typeIsPtr:   s.typeIsPtr,
		typeName:    s.typeName,
		typePkgPath: s.typePkgPath,
		typeValue:   s.typeValue,
		typeDepend:  s.typeDepend,
	}
}

func getStructType(val reflect.Type, isPtr bool) (ret model.FieldType, err error) {
	tVal, tErr := util.GetTypeValueEnum(val)
	if tErr != nil {
		err = tErr
		return
	}

	if util.IsStructType(tVal) {
		ret = &typeStruct{typeValue: tVal, typeName: val.String(), typePkgPath: val.PkgPath(), typeIsPtr: isPtr, typeDepend: val}
		return
	}

	err = fmt.Errorf("illegal struct type, type:%s", val.String())

	return
}
