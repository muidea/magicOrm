package common

import (
	"encoding/json"
	"fmt"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

type Common struct {
	entityModel   model.Model
	modelProvider provider.Provider
	specialPrefix string

	// temp value, for performance optimization
	entityTableName string
	entityValue     string
}

func New(vModel model.Model, modelProvider provider.Provider, prefix string) Common {
	return Common{entityModel: vModel, modelProvider: modelProvider, specialPrefix: prefix}
}

func (s *Common) constructTableName(vModel model.Model) string {
	strName := vModel.GetName()
	return strings.ToUpper(strName[:1]) + strName[1:]
}

func (s *Common) constructInfix(vFiled model.Field) string {
	strName := vFiled.GetName()
	return strings.ToUpper(strName[:1]) + strName[1:]
}

func (s *Common) GetTableName() string {
	if s.entityTableName == "" {
		s.entityTableName = s.GetHostTableName(s.entityModel)
	}

	return s.entityTableName
}

func (s *Common) GetHostTableName(vModel model.Model) string {
	tableName := s.constructTableName(vModel)
	if s.specialPrefix != "" {
		tableName = fmt.Sprintf("%s_%s", s.specialPrefix, tableName)
	}

	return tableName
}

func (s *Common) GetRelationTableName(vField model.Field, rModel model.Model) string {
	leftName := s.constructTableName(s.entityModel)
	rightName := s.constructTableName(rModel)
	infixVal := s.constructInfix(vField)

	tableName := fmt.Sprintf("%s%s%s%s", leftName, infixVal, getFieldRelation(vField), rightName)
	if s.specialPrefix != "" {
		tableName = fmt.Sprintf("%s_%s", s.specialPrefix, tableName)
	}

	return tableName
}

func (s *Common) GetPrimaryKeyField(vModel model.Model) model.Field {
	if vModel == nil {
		return s.entityModel.GetPrimaryField()
	}

	return vModel.GetPrimaryField()
}

func (s *Common) GetFields() model.Fields {
	return s.entityModel.GetFields()
}

func (s *Common) GetModelValue() (ret string, err *cd.Result) {
	if s.entityValue == "" {
		entityVal, entityErr := s.BuildModelValue(s.entityModel)
		if entityErr != nil {
			err = entityErr
			return
		}

		s.entityValue = entityVal
	}

	ret = s.entityValue
	return
}

func (s *Common) GetRelationValue(rModel model.Model) (leftVal, rightVal string, err *cd.Result) {
	entityVal, entityErr := s.GetModelValue()
	if entityErr != nil {
		err = entityErr
		log.Errorf("GetRelationValue failed, s.GetModelValue error:%s", err.Error())
		return
	}

	relationVal, relationErr := s.BuildModelValue(rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("GetRelationValue failed, s.EncodeModelValue error:%s", err.Error())
		return
	}

	leftVal = entityVal
	rightVal = relationVal
	return
}

func (s *Common) EncodeValue(vType model.Type, vValue model.Value) (ret interface{}, err *cd.Result) {
	defer func() {
		if eErr := recover(); eErr != nil {
			err = cd.NewError(cd.UnExpected, fmt.Sprintf("encode value failed, type:%v, err:%v", vType.GetPkgKey(), eErr))
			log.Errorf("EncodeValue failed, error:%s", err.Error())
		}
	}()

	eVal, eErr := s.modelProvider.EncodeValue(vValue, vType)
	if eErr != nil {
		err = eErr
		log.Errorf("EncodeValue failed, s.modelProvider.EncodeValue error:%s", err.Error())
		return
	}

	ret = eVal
	return
}

func (s *Common) GetTypeModel(vType model.Type) (ret model.Model, err *cd.Result) {
	vModel, vErr := s.modelProvider.GetTypeModel(vType)
	if vErr != nil {
		err = vErr
		log.Errorf("GetTypeModel failed, s.modelProvider.GetTypeModel error:%s", err.Error())
		return
	}

	ret = vModel
	return
}

func (s *Common) encodeStringValue(vType model.Type, vValue model.Value) (ret string, err *cd.Result) {
	fEncodeVal, fEncodeErr := s.EncodeValue(vType, vValue)
	if fEncodeErr != nil {
		err = fEncodeErr
		log.Errorf("encodeStringValue failed, s.EncodeValue error:%s", err.Error())
		return
	}
	ret = fEncodeVal.(string)
	return
}

func (s *Common) encodeIntValue(vType model.Type, vValue model.Value) (ret string, err *cd.Result) {
	fEncodeVal, fEncodeErr := s.EncodeValue(vType, vValue)
	if fEncodeErr != nil {
		err = fEncodeErr
		log.Errorf("encodeIntValue failed, s.EncodeValue error:%s", err.Error())
		return
	}
	ret = fmt.Sprintf("%v", fEncodeVal)
	return
}

func (s *Common) encodeFloatValue(vType model.Type, vValue model.Value) (ret string, err *cd.Result) {
	fEncodeVal, fEncodeErr := s.EncodeValue(vType, vValue)
	if fEncodeErr != nil {
		err = fEncodeErr
		log.Errorf("encodeFloatValue failed, s.EncodeValue error:%s", err.Error())
		return
	}
	ret = fmt.Sprintf("%v", fEncodeVal)
	return
}

func (s *Common) encodeSliceString(sliceVal []interface{}) []string {
	strSlice := make([]string, len(sliceVal))
	for idx, val := range sliceVal {
		strSlice[idx] = val.(string)
	}

	return strSlice
}

func (s *Common) encodeSliceInt(sliceVal []interface{}) []string {
	strSlice := make([]string, len(sliceVal))
	for idx, val := range sliceVal {
		strSlice[idx] = fmt.Sprintf("%v", val)
	}

	return strSlice
}

func (s *Common) encodeSliceFloat(sliceVal []interface{}) []string {
	strSlice := make([]string, len(sliceVal))
	for idx, val := range sliceVal {
		strSlice[idx] = fmt.Sprintf("%v", val)
	}

	return strSlice
}

func (s *Common) encodeSliceValue(vType model.Type, vValue model.Value) (ret []string, err *cd.Result) {
	fEncodeVal, fEncodeErr := s.EncodeValue(vType, vValue)
	if fEncodeErr != nil {
		err = fEncodeErr
		log.Errorf("encodeSliceValue failed, s.EncodeValue error:%s", err.Error())
		return
	}

	sliceVal, sliceOK := fEncodeVal.([]interface{})
	if !sliceOK {
		err = cd.NewError(cd.UnExpected, "illegal slice encode value")
		return
	}

	switch vType.Elem().GetValue() {
	case model.TypeStringValue, model.TypeDateTimeValue:
		ret = s.encodeSliceString(sliceVal)
	case model.TypeBooleanValue, model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeBigIntegerValue, model.TypeIntegerValue,
		model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveBigIntegerValue, model.TypePositiveIntegerValue:
		ret = s.encodeSliceInt(sliceVal)
	case model.TypeFloatValue, model.TypeDoubleValue:
		ret = s.encodeSliceFloat(sliceVal)
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal filed type %s", vType.Elem().GetPkgKey()))
	}

	return
}

func (s *Common) BuildModelValue(vModel model.Model) (ret string, err *cd.Result) {
	pkField := vModel.GetPrimaryField()
	switch pkField.GetType().GetValue() {
	case model.TypeStringValue:
		strVal, strErr := s.encodeStringValue(pkField.GetType(), pkField.GetValue())
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strVal)
	case model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeBigIntegerValue, model.TypeIntegerValue,
		model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveBigIntegerValue, model.TypePositiveIntegerValue:
		ret, err = s.encodeIntValue(pkField.GetType(), pkField.GetValue())
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal pkFiled type %s", pkField.GetType().GetPkgKey()))
	}
	return
}

func (s *Common) BuildFieldValue(vType model.Type, vValue model.Value) (ret string, err *cd.Result) {
	if vValue.IsNil() {
		ret, err = getTypeDefaultValue(vType)
		return
	}

	switch vType.GetValue() {
	case model.TypeStringValue, model.TypeDateTimeValue:
		strVal, strErr := s.encodeStringValue(vType, vValue)
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strVal)
	case model.TypeBooleanValue, model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeBigIntegerValue, model.TypeIntegerValue,
		model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveBigIntegerValue, model.TypePositiveIntegerValue:
		ret, err = s.encodeIntValue(vType, vValue)
	case model.TypeFloatValue, model.TypeDoubleValue:
		ret, err = s.encodeFloatValue(vType, vValue)
	case model.TypeSliceValue:
		fEncodeVal, fEncodeErr := s.EncodeValue(vType, vValue)
		if fEncodeErr != nil {
			err = fEncodeErr
			log.Errorf("encodeIntValue failed, s.EncodeValue error:%s", err.Error())
			return
		}
		byteVal, byteErr := json.Marshal(fEncodeVal)
		if byteErr != nil {
			err = cd.NewError(cd.UnExpected, fmt.Sprintf("%s", byteErr.Error()))
			return
		}
		ret = fmt.Sprintf("'%v'", strings.ReplaceAll(string(byteVal), "'", "''"))
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal filed type %s", vType.GetPkgKey()))
	}
	return
}

func (s *Common) BuildOprValue(vType model.Type, vValue model.Value) (ret string, err *cd.Result) {
	if vValue.IsNil() {
		err = cd.NewError(cd.UnExpected, "nil opr value")
		return
	}

	switch vType.GetValue() {
	case model.TypeStringValue, model.TypeDateTimeValue:
		strVal, strErr := s.encodeStringValue(vType, vValue)
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strVal)
	case model.TypeBooleanValue, model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeBigIntegerValue, model.TypeIntegerValue,
		model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveBigIntegerValue, model.TypePositiveIntegerValue:
		ret, err = s.encodeIntValue(vType, vValue)
	case model.TypeFloatValue, model.TypeDoubleValue:
		ret, err = s.encodeFloatValue(vType, vValue)
	case model.TypeSliceValue:
		sliceVal, sliceErr := s.encodeSliceValue(vType, vValue)
		if sliceErr != nil {
			err = sliceErr
		}
		ret = strings.Join(sliceVal, ",")
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal filed type %s", vType.GetPkgKey()))
	}
	return
}
