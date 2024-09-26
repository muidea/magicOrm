package common

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/model"
)

type relationType int

const (
	relationInvalid = 0
	relationHas1v1  = 1
	relationHas1vn  = 2
	relationRef1v1  = 3
	relationRef1vn  = 4
)

func (s relationType) String() string {
	return fmt.Sprintf("%d", s)
}

func getFieldRelation(vField model.Field) (ret relationType) {
	fType := vField.GetType()
	if fType.IsBasic() {
		return
	}

	isPtr := fType.Elem().IsPtrType() || fType.IsPtrType()
	isSlice := model.IsSliceType(fType.GetValue())

	if !isPtr && !isSlice {
		ret = relationHas1v1
		return
	}

	if !isPtr && isSlice {
		ret = relationHas1vn
		return
	}

	if isPtr && !isSlice {
		ret = relationRef1v1
		return
	}

	ret = relationRef1vn
	return
}

func getTypeDefaultValue(fType model.Type) (ret string, err *cd.Result) {
	switch fType.GetValue() {
	case model.TypeBooleanValue, model.TypeBitValue,
		model.TypeSmallIntegerValue, model.TypePositiveBitValue,
		model.TypeIntegerValue, model.TypeInteger32Value, model.TypePositiveSmallIntegerValue,
		model.TypeBigIntegerValue, model.TypePositiveIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveBigIntegerValue,
		model.TypeFloatValue, model.TypeDoubleValue:
		ret = "0"
		break
	case model.TypeStringValue,
		model.TypeDateTimeValue,
		model.TypeSliceValue:
		ret = "''"
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("no support field type, type:%v", fType.GetPkgKey()))
	}

	if err != nil {
		log.Errorf("getTypeDefaultValue failed, error:%s", err.Error())
	}

	return
}
