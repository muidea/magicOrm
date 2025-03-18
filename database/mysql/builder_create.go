package mysql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

func (s *Builder) BuildCreateTable(vModel model.Model) (ret *ResultStack, err *cd.Result) {
	createSQL := ""
	for _, field := range vModel.GetFields() {
		if !model.IsBasicField(field) {
			continue
		}

		infoVal, infoErr := s.declareFieldInfo(field)
		if infoErr != nil {
			err = infoErr
			log.Errorf("BuildCreateTable failed, declareFieldInfo error:%s", err.Error())
			return
		}

		if createSQL == "" {
			createSQL = fmt.Sprintf("\t%s", infoVal)
		} else {
			createSQL = fmt.Sprintf("%s,\n\t%s", createSQL, infoVal)
		}
	}

	pkFieldName := vModel.GetPrimaryField().GetName()
	createSQL = fmt.Sprintf("%s,\n\tPRIMARY KEY (`%s`)", createSQL, pkFieldName)

	createSQL = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n%s\n)\n", s.buildCodec.ConstructModelTableName(vModel), createSQL)
	if traceSQL() {
		log.Infof("[SQL] create: %s", createSQL)
	}

	ret = NewResult(createSQL, nil)
	return
}

// BuildCreateRelationTable Build CreateRelation Schema
func (s *Builder) BuildCreateRelationTable(vModel model.Model, vField model.Field) (ret *ResultStack, err *cd.Result) {
	lPKField := vModel.GetPrimaryField()
	lPKType, lPKErr := getTypeDeclare(lPKField.GetType(), lPKField.GetSpec())
	if lPKErr != nil {
		err = lPKErr
		log.Errorf("BuildCreateRelationTable failed, getTypeDeclare error:%s", err.Error())
		return
	}

	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildCreateRelationTable %s failed, s.buildCodec.ConstructRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType().Elem())
	if rErr != nil {
		err = rErr
		log.Errorf("BuildCreateRelationTable %s failed, s.modelProvider.GetTypeModel error:%s", vField.GetName(), err.Error())
		return
	}

	rPKField := rModel.GetPrimaryField()
	rPKType, rPKErr := getTypeDeclare(rPKField.GetType(), rPKField.GetSpec())
	if rPKErr != nil {
		err = rPKErr
		log.Errorf("BuildCreateRelationTable failed, getTypeDeclare error:%s", err.Error())
		return
	}

	createRelationSQL := fmt.Sprintf("\t`id` BIGINT NOT NULL AUTO_INCREMENT,\n\t`left` %s NOT NULL,\n\t`right` %s NOT NULL,\n\tPRIMARY KEY (`id`),\n\tINDEX(`left`)", lPKType, rPKType)
	createRelationSQL = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n%s\n)\n", relationTableName, createRelationSQL)
	if traceSQL() {
		log.Infof("[SQL] create relation: %s", createRelationSQL)
	}

	ret = NewResult(createRelationSQL, nil)
	return
}

// declareFieldInfo declare field info
// 根据字段类型和字段特性生成字段定义
// 类似以下信息
// `id` int(11) NOT NULL AUTO_INCREMENT
// `i8` tinyint(4) DEFAULT '100',
func (s *Builder) declareFieldInfo(vField model.Field) (ret string, err *cd.Result) {
	strBuffer := bytes.NewBufferString("")
	// Write field name
	strBuffer.WriteString("`")
	strBuffer.WriteString(vField.GetName())
	strBuffer.WriteString("`")

	// Write field type
	typeVal, typeErr := getTypeDeclare(vField.GetType(), vField.GetSpec())
	if typeErr != nil {
		err = typeErr
		log.Errorf("declareFieldInfo failed, getTypeDeclare error:%s", err.Error())
		return
	}
	strBuffer.WriteString(" ")
	strBuffer.WriteString(typeVal)

	// Write NULL constraint
	if !vField.GetType().IsPtrType() {
		strBuffer.WriteString(" NOT NULL")
	}

	// Write default value if exists
	fSpec := vField.GetSpec()
	if fSpec != nil && fSpec.GetDefaultValue() != nil {
		defaultValue, defaultErr := s.validDefaultValue(vField.GetType(), fSpec)
		if defaultErr != nil {
			err = defaultErr
			log.Errorf("declareFieldInfo failed, validDefaultValue error:%s", err.Error())
			return
		}
		if defaultValue != "" {
			strBuffer.WriteString(" DEFAULT ")
			strBuffer.WriteString(defaultValue)
		}
	}

	// Write auto increment if needed
	if fSpec != nil && model.IsAutoIncrementDeclare(fSpec.GetValueDeclare()) {
		strBuffer.WriteString(" AUTO_INCREMENT")
	}

	ret = strBuffer.String()
	return
}

func (s *Builder) validDefaultValue(vType model.Type, vSpec model.Spec) (ret string, err *cd.Result) {
	if !model.IsBasic(vType) || vType.IsPtrType() {
		err = cd.NewResult(cd.UnExpected, "default value must be basic type and not ptr type")
		return
	}

	var defaultValue any
	defaultValueDeclare := vSpec.GetDefaultValue()
	if defaultValueDeclare != nil {
		switch val := defaultValueDeclare.(type) {
		case string:
			if strings.Contains(val, "$referenceValue.") {
				vTypeDefaultVal, _ := vType.Interface(nil)
				defaultValue = vTypeDefaultVal.Get()
			} else {
				defaultValue = val
			}
		default:
			byteVal, byteErr := json.Marshal(val)
			if byteErr != nil {
				err = cd.NewResult(cd.UnExpected, byteErr.Error())
				return
			}
			defaultValue = string(byteVal)
		}
	} else {
		vTypeDefaultVal, _ := vType.Interface(nil)
		defaultValue = vTypeDefaultVal.Get()
	}

	switch vType.Elem().GetValue() {
	case model.TypeStringValue, model.TypeDateTimeValue:
		ret = fmt.Sprintf("'%v'", defaultValue)
	case model.TypeBooleanValue, model.TypeBitValue,
		model.TypeSmallIntegerValue, model.TypePositiveBitValue,
		model.TypeIntegerValue, model.TypeInteger32Value, model.TypePositiveSmallIntegerValue,
		model.TypeBigIntegerValue, model.TypePositiveIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveBigIntegerValue,
		model.TypeFloatValue, model.TypeDoubleValue:
		ret = fmt.Sprintf("%v", defaultValue)
	}

	return
}
