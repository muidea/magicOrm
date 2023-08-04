package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

func verifyField(vField model.Field) error {
	fName := vField.GetName()
	if IsKeyWord(fName) {
		return fmt.Errorf("illegal fieldSpec, is a key word.[%s]", fName)
	}

	return nil
}

func verifyModel(vModel model.Model) error {
	name := vModel.GetName()
	if IsKeyWord(name) {
		return fmt.Errorf("illegal structName, is a key word.[%s]", name)
	}

	for _, val := range vModel.GetFields() {
		err := verifyField(val)
		if err != nil {
			return err
		}
	}

	return nil
}

func declareFieldInfo(vField model.Field) (ret string, err error) {
	autoIncrement := ""
	fSpec := vField.GetSpec()
	if fSpec != nil && fSpec.IsAutoIncrement() {
		autoIncrement = "AUTO_INCREMENT"
	}

	allowNull := "NOT NULL"
	fType := vField.GetType()
	if fType.IsPtrType() {
		allowNull = ""
	}

	typeVal, typeErr := getFieldType(vField)
	if typeErr != nil {
		err = typeErr
		return
	}

	ret = fmt.Sprintf("`%s` %s %s %s", vField.GetName(), typeVal, allowNull, autoIncrement)
	return
}

func getFieldType(info model.Field) (ret string, err error) {
	fType := info.GetType()
	switch fType.GetValue() {
	case model.TypeBooleanValue, model.TypeBitValue:
		ret = "TINYINT"
		break
	case model.TypeStringValue:
		ret = "TEXT"
		break
	case model.TypeDateTimeValue:
		ret = "DATETIME"
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
		err = fmt.Errorf("no support fileType, name:%s, type:%d", info.GetName(), fType.GetValue())
	}

	return
}
