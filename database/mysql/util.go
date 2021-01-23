package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func verifyField(info model.Field) error {
	fTag := info.GetTag()
	if IsKeyWord(fTag.GetName()) {
		return fmt.Errorf("illegal fieldTag, is a key word.[%s]", fTag)
	}

	return nil
}

func verifyModel(info model.Model) error {
	name := info.GetName()
	if IsKeyWord(name) {
		return fmt.Errorf("illegal structName, is a key word.[%s]", name)
	}

	for _, val := range info.GetFields() {
		err := verifyField(val)
		if err != nil {
			return err
		}
	}

	return nil
}

func declareFieldInfo(info model.Field) (ret string, err error) {
	autoIncrement := ""
	fTag := info.GetTag()
	if fTag.IsAutoIncrement() {
		autoIncrement = "AUTO_INCREMENT"
	}

	allowNull := "NOT NULL"
	fType := info.GetType()
	if fType.IsPtrType() {
		allowNull = ""
	}

	infoVal, infoErr := getFieldType(info)
	if infoErr != nil {
		err = infoErr
		return
	}

	ret = fmt.Sprintf("`%s` %s %s %s", fTag.GetName(), infoVal, allowNull, autoIncrement)
	return
}

func getFieldType(info model.Field) (ret string, err error) {
	fType := info.GetType()
	switch fType.GetValue() {
	case util.TypeBooleanField:
		ret = "TINYINT"
		break
	case util.TypeStringField:
		ret = "TEXT"
		break
	case util.TypeDateTimeField:
		ret = "DATETIME"
		break
	case util.TypeBitField:
		ret = "TINYINT"
		break
	case util.TypeSmallIntegerField:
		ret = "SMALLINT"
		break
	case util.TypeIntegerField:
		ret = "INT"
		break
	case util.TypeInteger32Field:
		ret = "INT"
		break
	case util.TypeBigIntegerField:
		ret = "BIGINT"
		break
	case util.TypePositiveBitField:
		ret = "SMALLINT"
		break
	case util.TypePositiveSmallIntegerField:
		ret = "INT"
		break
	case util.TypePositiveIntegerField:
		ret = "BIGINT"
		break
	case util.TypePositiveInteger32Field:
		ret = "BIGINT"
		break
	case util.TypePositiveBigIntegerField:
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

func getFieldInitializeValue(field model.Field) (ret interface{}, err error) {
	fType := field.GetType()
	switch fType.GetValue() {
	case util.TypeBooleanField:
		val := int8(0)
		ret = &val
		break
	case util.TypeBitField:
		val := int8(0)
		ret = &val
		break
	case util.TypeSmallIntegerField:
		val := int16(0)
		ret = &val
		break
	case util.TypeIntegerField:
		val := int(0)
		ret = &val
		break
	case util.TypeInteger32Field:
		val := int32(0)
		ret = &val
		break
	case util.TypeBigIntegerField:
		val := int64(0)
		ret = &val
		break
	case util.TypePositiveBitField:
		val := uint8(0)
		ret = &val
		break
	case util.TypePositiveSmallIntegerField:
		val := uint16(0)
		ret = &val
		break
	case util.TypePositiveIntegerField:
		val := uint(0)
		ret = &val
		break
	case util.TypePositiveInteger32Field:
		val := uint32(0)
		ret = &val
		break
	case util.TypePositiveBigIntegerField:
		val := uint64(0)
		ret = &val
		break
	case util.TypeFloatField:
		val := float32(0.00)
		ret = &val
		break
	case util.TypeDoubleField:
		val := 0.0000
		ret = &val
		break
	case util.TypeStringField, util.TypeDateTimeField:
		val := ""
		ret = &val
		break
	case util.TypeSliceField:
		if fType.IsBasic() {
			val := ""
			ret = &val
		} else {
			err = fmt.Errorf("no support fileType, name:%s, type:%d", field.GetName(), fType.GetValue())
		}
	default:
		err = fmt.Errorf("no support fileType, name:%s, type:%d", field.GetName(), fType.GetValue())
	}

	return
}
