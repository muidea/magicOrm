package postgres

import (
	"fmt"
	"os"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/models"
)

func traceSQL() bool {
	enableTrace, enableOK := os.LookupEnv("TRACE_SQL")
	if enableOK && enableTrace == "true" {
		return true
	}

	return false
}

func getTypeDeclare(fType models.Type, fSpec models.Spec, atPKField bool) (ret string, err *cd.Error) {
	// 检查是否为自增字段
	isAutoIncrement := false
	if fSpec != nil && fSpec.IsPrimaryKey() && fType.GetValue().IsNumberValueType() && atPKField {
		isAutoIncrement = models.IsAutoIncrementDeclare(fSpec.GetValueDeclare())
	}

	switch fType.GetValue() {
	case models.TypeStringValue:
		if fSpec != nil && fSpec.IsPrimaryKey() {
			ret = "VARCHAR(32)"
		} else {
			ret = "TEXT"
		}
	case models.TypeDateTimeValue:
		ret = "TIMESTAMP(3)"
	case models.TypeBooleanValue:
		ret = "BOOLEAN"
	case models.TypeByteValue:
		ret = "SMALLINT"
	case models.TypeSmallIntegerValue, models.TypePositiveByteValue:
		if isAutoIncrement {
			ret = "SMALLSERIAL"
		} else {
			ret = "SMALLINT"
		}
	case models.TypeIntegerValue, models.TypeInteger32Value, models.TypePositiveSmallIntegerValue:
		if isAutoIncrement {
			ret = "SERIAL"
		} else {
			ret = "INTEGER"
		}
	case models.TypeBigIntegerValue, models.TypePositiveIntegerValue, models.TypePositiveInteger32Value, models.TypePositiveBigIntegerValue:
		if isAutoIncrement {
			ret = "BIGSERIAL"
		} else {
			ret = "BIGINT"
		}
	case models.TypeFloatValue:
		ret = "REAL"
	case models.TypeDoubleValue:
		ret = "DOUBLE PRECISION"
	case models.TypeSliceValue:
		ret = "TEXT"
	default:
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("no support field type, type:%v", fType.GetPkgKey()))
	}

	if err != nil {
		log.Errorf("getTypeDeclare failed, error:%s", err.Error())
	}

	return
}

func getFieldValueHolder(fType models.Type) (ret any, err *cd.Error) {
	switch fType.GetValue() {
	case models.TypeStringValue, models.TypeDateTimeValue:
		ret = ""
	case models.TypeBooleanValue:
		ret = false
	case models.TypeByteValue:
		ret = int8(0)
	case models.TypeSmallIntegerValue:
		ret = int16(0)
	case models.TypeIntegerValue:
		ret = int(0)
	case models.TypeInteger32Value:
		ret = int32(0)
	case models.TypeBigIntegerValue:
		ret = int64(0)
	case models.TypePositiveByteValue:
		ret = uint16(0)
	case models.TypePositiveSmallIntegerValue:
		ret = uint32(0)
	case models.TypePositiveIntegerValue:
		ret = uint(0)
	case models.TypePositiveInteger32Value, models.TypePositiveBigIntegerValue:
		ret = uint64(0)
	case models.TypeFloatValue:
		ret = float32(0.00)
	case models.TypeDoubleValue:
		ret = 0.0000
	case models.TypeSliceValue:
		if models.IsBasic(fType.Elem()) {
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
