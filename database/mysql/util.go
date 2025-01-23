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

func verifyField(vField model.Field) *cd.Result {
	fName := vField.GetName()
	if IsKeyWord(fName) {
		return cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal fieldSpec, is a key word.[%s]", fName))
	}

	return nil
}

func verifyModel(vModel model.Model) *cd.Result {
	name := vModel.GetName()
	if IsKeyWord(name) {
		return cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal structName, is a key word.[%s]", name))
	}

	for _, val := range vModel.GetFields() {
		err := verifyField(val)
		if err != nil {
			log.Errorf("verifyModel failed, verifyField error:%s", err.Error())
			return err
		}
	}

	return nil
}

func getTypeDeclare(fType model.Type, fSpec model.Spec) (ret string, err *cd.Result) {
	switch fType.GetValue() {
	case model.TypeStringValue:
		if fSpec.IsPrimaryKey() {
			ret = "VARCHAR(32)"
		} else {
			ret = "TEXT"
		}
		break
	case model.TypeDateTimeValue:
		ret = "DATETIME"
		break
	case model.TypeBooleanValue, model.TypeBitValue:
		ret = "TINYINT"
		break
	case model.TypeSmallIntegerValue, model.TypePositiveBitValue:
		ret = "SMALLINT"
		break
	case model.TypeIntegerValue, model.TypeInteger32Value, model.TypePositiveSmallIntegerValue:
		ret = "INT"
		break
	case model.TypeBigIntegerValue, model.TypePositiveIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveBigIntegerValue:
		ret = "BIGINT"
		break
	case model.TypeFloatValue:
		ret = "FLOAT"
		break
	case model.TypeDoubleValue:
		ret = "DOUBLE"
		break
	case model.TypeSliceValue:
		ret = "TEXT"
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("no support field type, type:%v", fType.GetPkgKey()))
	}

	if err != nil {
		log.Errorf("getTypeDeclare failed, error:%s", err.Error())
	}

	return
}

func getFieldPlaceHolder(field model.Field) (ret interface{}, err *cd.Result) {
	fType := field.GetType()
	switch fType.GetValue() {
	case model.TypeStringValue, model.TypeDateTimeValue:
		val := ""
		ret = &val
		break
	case model.TypeBooleanValue, model.TypeBitValue:
		val := int8(0)
		ret = &val
		break
	case model.TypeSmallIntegerValue:
		val := int16(0)
		ret = &val
		break
	case model.TypeIntegerValue:
		val := int(0)
		ret = &val
		break
	case model.TypeInteger32Value:
		val := int32(0)
		ret = &val
		break
	case model.TypeBigIntegerValue:
		val := int64(0)
		ret = &val
		break
	case model.TypePositiveBitValue:
		val := uint16(0)
		ret = &val
		break
	case model.TypePositiveSmallIntegerValue:
		val := uint32(0)
		ret = &val
		break
	case model.TypePositiveIntegerValue:
		val := uint(0)
		ret = &val
		break
	case model.TypePositiveInteger32Value, model.TypePositiveBigIntegerValue:
		val := uint64(0)
		ret = &val
		break
	case model.TypeFloatValue:
		val := float32(0.00)
		ret = &val
		break
	case model.TypeDoubleValue:
		val := 0.0000
		ret = &val
		break
	case model.TypeSliceValue:
		if fType.IsBasic() {
			val := ""
			ret = &val
		} else {
			err = cd.NewResult(cd.UnExpected, fmt.Sprintf("no support fileType, name:%s, type:%v", field.GetName(), fType.GetPkgKey()))
		}
	default:
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("no support fileType, name:%s, type:%v", field.GetName(), fType.GetPkgKey()))
	}

	if err != nil {
		log.Errorf("getFieldPlaceHolder failed, error:%s", err.Error())
	}

	return
}
