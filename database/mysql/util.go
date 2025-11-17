package mysql

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

func getTypeDeclare(fType models.Type, fSpec models.Spec) (ret string, err *cd.Error) {
	switch fType.GetValue() {
	case models.TypeStringValue:
		if fSpec.IsPrimaryKey() {
			ret = "VARCHAR(32)"
		} else {
			ret = "TEXT"
		}
	case models.TypeDateTimeValue:
		ret = "DATETIME(3)"
	case models.TypeBooleanValue, models.TypeByteValue:
		ret = "TINYINT"
	case models.TypeSmallIntegerValue, models.TypePositiveByteValue:
		ret = "SMALLINT"
	case models.TypeIntegerValue, models.TypeInteger32Value, models.TypePositiveSmallIntegerValue:
		ret = "INT"
	case models.TypeBigIntegerValue, models.TypePositiveIntegerValue, models.TypePositiveInteger32Value, models.TypePositiveBigIntegerValue:
		ret = "BIGINT"
	case models.TypeFloatValue:
		ret = "FLOAT"
	case models.TypeDoubleValue:
		ret = "DOUBLE"
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

func getFieldPlaceHolder(fType models.Type) (ret interface{}, err *cd.Error) {
	switch fType.GetValue() {
	case models.TypeStringValue, models.TypeDateTimeValue:
		val := ""
		ret = &val
	case models.TypeBooleanValue, models.TypeByteValue:
		val := int8(0)
		ret = &val
	case models.TypeSmallIntegerValue:
		val := int16(0)
		ret = &val
	case models.TypeIntegerValue:
		val := int(0)
		ret = &val
	case models.TypeInteger32Value:
		val := int32(0)
		ret = &val
	case models.TypeBigIntegerValue:
		val := int64(0)
		ret = &val
	case models.TypePositiveByteValue:
		val := uint16(0)
		ret = &val
	case models.TypePositiveSmallIntegerValue:
		val := uint32(0)
		ret = &val
	case models.TypePositiveIntegerValue:
		val := uint(0)
		ret = &val
	case models.TypePositiveInteger32Value, models.TypePositiveBigIntegerValue:
		val := uint64(0)
		ret = &val
	case models.TypeFloatValue:
		val := float32(0.00)
		ret = &val
	case models.TypeDoubleValue:
		val := 0.0000
		ret = &val
	case models.TypeSliceValue:
		if models.IsBasic(fType.Elem()) {
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
