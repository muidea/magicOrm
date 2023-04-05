package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func verifyField(vField model.Field) error {
	fTag := vField.GetTag()
	if IsKeyWord(fTag.GetName()) {
		return fmt.Errorf("illegal fieldTag, is a key word.[%s]", fTag)
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
	fTag := vField.GetTag()
	if fTag.IsAutoIncrement() {
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

	ret = fmt.Sprintf("`%s` %s %s %s", fTag.GetName(), typeVal, allowNull, autoIncrement)
	return
}

func getFieldType(info model.Field) (ret string, err error) {
	fType := info.GetType()
	switch fType.GetValue() {
	case util.TypeBooleanField, util.TypeBitField:
		ret = "TINYINT"
		break
	case util.TypeStringField:
		ret = "TEXT"
		break
	case util.TypeDateTimeField:
		ret = "DATETIME"
		break
	case util.TypeSmallIntegerField, util.TypePositiveBitField:
		ret = "SMALLINT"
		break
	case util.TypeIntegerField, util.TypeInteger32Field, util.TypePositiveSmallIntegerField:
		ret = "INT"
		break
	case util.TypeBigIntegerField, util.TypePositiveIntegerField, util.TypePositiveInteger32Field, util.TypePositiveBigIntegerField:
		ret = "BIGINT"
		break
	case util.TypeFloatField:
		ret = "FLOAT"
		break
	case util.TypeDoubleField:
		ret = "DOUBLE"
		break
	case util.TypeSliceField:
		ret = "TEXT"
	default:
		err = fmt.Errorf("no support fileType, name:%s, type:%d", info.GetName(), fType.GetValue())
	}

	return
}
