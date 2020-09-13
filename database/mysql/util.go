package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func verifyFieldInfo(fieldInfo model.Field) error {
	fTag := fieldInfo.GetTag()
	if IsKeyWord(fTag.GetName()) {
		return fmt.Errorf("illegal fieldTag, is a key word.[%s]", fTag)
	}

	return nil
}

func verifyModelInfo(modelInfo model.Model) error {
	name := modelInfo.GetName()
	if IsKeyWord(name) {
		return fmt.Errorf("illegal structName, is a key word.[%s]", name)
	}

	for _, val := range modelInfo.GetFields() {
		err := verifyFieldInfo(val)
		if err != nil {
			return err
		}
	}

	return nil
}

func declareFieldInfo(fieldInfo model.Field) (ret string, err error) {
	autoIncrement := ""
	fTag := fieldInfo.GetTag()
	if fTag.IsAutoIncrement() {
		autoIncrement = "AUTO_INCREMENT"
	}

	allowNull := "NOT NULL"
	fType := fieldInfo.GetType()
	if fType.IsPtrType() {
		allowNull = ""
	}

	infoVal, infoErr := getFieldType(fieldInfo)
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

func getFieldInitValue(info model.Field) (ret interface{}, err error) {
	fType := info.GetType()
	if fType.Depend() != nil {
		fType = fType.Depend()
	}

	if !util.IsStructType(fType.GetValue()) {
		return
	}

	switch fType.GetValue() {
	case util.TypeBooleanField, util.TypeBitField:
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
	case util.TypeStringField, util.TypeDateTimeField, util.TypeSliceField:
		val := ""
		ret = &val
		break
	default:
		err = fmt.Errorf("no support fileType, name:%s, type:%d", info.GetName(), fType.GetValue())
	}

	return
}
