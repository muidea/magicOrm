package mysql

import (
	"fmt"
	"os"
	"sync/atomic"

	"log/slog"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
)

var traceSQLEnabled int32 = -1

func traceSQL() bool {
	if enabled := atomic.LoadInt32(&traceSQLEnabled); enabled >= 0 {
		return enabled == 1
	}

	enableTrace, enableOK := os.LookupEnv("TRACE_SQL")
	enabled := int32(0)
	if enableOK && enableTrace == "true" {
		enabled = 1
	}
	atomic.StoreInt32(&traceSQLEnabled, enabled)
	return enabled == 1
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
		slog.Error("getTypeDeclare failed", "error", err.Error())
	}

	return
}

func getFieldPlaceHolder(fType models.Type) (ret any, err *cd.Error) {
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
		slog.Error("getFieldPlaceHolder failed", "error", err.Error())
	}

	return
}
