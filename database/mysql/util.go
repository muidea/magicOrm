package mysql

import (
	"fmt"
	"os"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

func traceSQL() bool {
	enableTrace, enableOK := os.LookupEnv("TRACE_SQL")
	if enableOK && enableTrace == "true" {
		return true
	}

	return false
}

func getTypeDeclare(fType model.Type, fSpec model.Spec) (ret string, err *cd.Error) {
	switch fType.GetValue() {
	case model.TypeStringValue:
		if fSpec.IsPrimaryKey() {
			ret = "VARCHAR(32)"
		} else {
			ret = "TEXT"
		}
	case model.TypeDateTimeValue:
		ret = "DATETIME"
	case model.TypeBooleanValue, model.TypeBitValue:
		ret = "TINYINT"
	case model.TypeSmallIntegerValue, model.TypePositiveBitValue:
		ret = "SMALLINT"
	case model.TypeIntegerValue, model.TypeInteger32Value, model.TypePositiveSmallIntegerValue:
		ret = "INT"
	case model.TypeBigIntegerValue, model.TypePositiveIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveBigIntegerValue:
		ret = "BIGINT"
	case model.TypeFloatValue:
		ret = "FLOAT"
	case model.TypeDoubleValue:
		ret = "DOUBLE"
	case model.TypeSliceValue:
		ret = "TEXT"
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("no support field type, type:%v", fType.GetPkgKey()))
	}

	if err != nil {
		log.Errorf("getTypeDeclare failed, error:%s", err.Error())
	}

	return
}

func getFieldPlaceHolder(fType model.Type) (ret interface{}, err *cd.Error) {
	switch fType.GetValue() {
	case model.TypeStringValue, model.TypeDateTimeValue:
		val := ""
		ret = &val
	case model.TypeBooleanValue, model.TypeBitValue:
		val := int8(0)
		ret = &val
	case model.TypeSmallIntegerValue:
		val := int16(0)
		ret = &val
	case model.TypeIntegerValue:
		val := int(0)
		ret = &val
	case model.TypeInteger32Value:
		val := int32(0)
		ret = &val
	case model.TypeBigIntegerValue:
		val := int64(0)
		ret = &val
	case model.TypePositiveBitValue:
		val := uint16(0)
		ret = &val
	case model.TypePositiveSmallIntegerValue:
		val := uint32(0)
		ret = &val
	case model.TypePositiveIntegerValue:
		val := uint(0)
		ret = &val
	case model.TypePositiveInteger32Value, model.TypePositiveBigIntegerValue:
		val := uint64(0)
		ret = &val
	case model.TypeFloatValue:
		val := float32(0.00)
		ret = &val
	case model.TypeDoubleValue:
		val := 0.0000
		ret = &val
	case model.TypeSliceValue:
		if model.IsBasic(fType.Elem()) {
			val := ""
			ret = &val
		} else {
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("no support fileType, type:%v", fType.GetPkgKey()))
		}
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("no support fileType, type:%v", fType.GetPkgKey()))
	}

	if err != nil {
		log.Errorf("getFieldPlaceHolder failed, error:%s", err.Error())
	}

	return
}
