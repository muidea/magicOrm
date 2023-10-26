package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

func verifyField(vField model.Field) *cd.Result {
	fName := vField.GetName()
	if IsKeyWord(fName) {
		return cd.NewError(cd.UnExpected, fmt.Sprintf("illegal fieldSpec, is a key word.[%s]", fName))
	}

	return nil
}

func verifyModel(vModel model.Model) *cd.Result {
	name := vModel.GetName()
	if IsKeyWord(name) {
		return cd.NewError(cd.UnExpected, fmt.Sprintf("illegal structName, is a key word.[%s]", name))
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

func declareFieldInfo(vField model.Field) (ret string, err *cd.Result) {
	autoIncrement := ""
	fSpec := vField.GetSpec()
	if fSpec != nil && model.IsAutoIncrement(fSpec.GetValueDeclare()) {
		autoIncrement = "AUTO_INCREMENT"
	}

	allowNull := "NOT NULL"
	typeVal, typeErr := getTypeDeclare(vField.GetType(), vField.GetSpec())
	if typeErr != nil {
		err = typeErr
		log.Errorf("declareFieldInfo failed, getTypeDeclare error:%s", err.Error())
		return
	}

	ret = fmt.Sprintf("`%s` %s %s %s", vField.GetName(), typeVal, allowNull, autoIncrement)
	return
}

func getTypeDeclare(fType model.Type, fSpec model.Spec) (ret string, err *cd.Result) {
	switch fType.GetValue() {
	case model.TypeStringValue:
		if fSpec.GetValueDeclare() == model.UUID {
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
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("no support field type, type:%v", fType.GetPkgKey()))
	}

	if err != nil {
		log.Errorf("getTypeDeclare failed, error:%s", err.Error())
	}

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

func getFieldScanDestPtr(field model.Field) (ret interface{}, err *cd.Result) {
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
			err = cd.NewError(cd.UnExpected, fmt.Sprintf("no support fileType, name:%s, type:%v", field.GetName(), fType.GetPkgKey()))
		}
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("no support fileType, name:%s, type:%v", field.GetName(), fType.GetPkgKey()))
	}

	if err != nil {
		log.Errorf("getFieldScanDestPtr failed, error:%s", err.Error())
	}

	return
}
