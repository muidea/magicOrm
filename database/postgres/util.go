package postgres

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

func getTypeDeclare(fType model.Type, fSpec model.Spec, atPKField bool) (ret string, err *cd.Error) {
	// 检查是否为自增字段
	isAutoIncrement := false
	if fSpec != nil && fSpec.IsPrimaryKey() && fType.GetValue().IsNumberValueType() && atPKField {
		isAutoIncrement = model.IsAutoIncrementDeclare(fSpec.GetValueDeclare())
	}

	switch fType.GetValue() {
	case model.TypeStringValue:
		if fSpec != nil && fSpec.IsPrimaryKey() {
			ret = "VARCHAR(32)"
		} else {
			ret = "TEXT"
		}
	case model.TypeDateTimeValue:
		ret = "TIMESTAMP(3)"
	case model.TypeBooleanValue:
		ret = "BOOLEAN"
	case model.TypeBitValue:
		ret = "SMALLINT"
	case model.TypeSmallIntegerValue, model.TypePositiveBitValue:
		if isAutoIncrement {
			ret = "SMALLSERIAL"
		} else {
			ret = "SMALLINT"
		}
	case model.TypeIntegerValue, model.TypeInteger32Value, model.TypePositiveSmallIntegerValue:
		if isAutoIncrement {
			ret = "SERIAL"
		} else {
			ret = "INTEGER"
		}
	case model.TypeBigIntegerValue, model.TypePositiveIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveBigIntegerValue:
		if isAutoIncrement {
			ret = "BIGSERIAL"
		} else {
			ret = "BIGINT"
		}
	case model.TypeFloatValue:
		ret = "REAL"
	case model.TypeDoubleValue:
		ret = "DOUBLE PRECISION"
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

func getFieldValueHolder(fType model.Type) (ret any, err *cd.Error) {
	switch fType.GetValue() {
	case model.TypeStringValue, model.TypeDateTimeValue:
		ret = ""
	case model.TypeBooleanValue:
		ret = false
	case model.TypeBitValue:
		ret = int8(0)
	case model.TypeSmallIntegerValue:
		ret = int16(0)
	case model.TypeIntegerValue:
		ret = int(0)
	case model.TypeInteger32Value:
		ret = int32(0)
	case model.TypeBigIntegerValue:
		ret = int64(0)
	case model.TypePositiveBitValue:
		ret = uint16(0)
	case model.TypePositiveSmallIntegerValue:
		ret = uint32(0)
	case model.TypePositiveIntegerValue:
		ret = uint(0)
	case model.TypePositiveInteger32Value, model.TypePositiveBigIntegerValue:
		ret = uint64(0)
	case model.TypeFloatValue:
		ret = float32(0.00)
	case model.TypeDoubleValue:
		ret = 0.0000
	case model.TypeSliceValue:
		if model.IsBasic(fType.Elem()) {
			ret = ""
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
