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
	hostModel     model.Model
	modelProvider provider.Provider
	specialPrefix string

	// temp value, for performance optimization
	hostModelTableName string
	hostModelValue     string
}

func New(vModel model.Model, modelProvider provider.Provider, prefix string) *Common {
	return &Common{hostModel: vModel, modelProvider: modelProvider, specialPrefix: prefix}
}

func (s *Common) constructTableName(vModel model.Model) string {
	strName := vModel.GetName()
	return strings.ToUpper(strName[:1]) + strName[1:]
}

func (s *Common) constructInfix(vFiled model.Field) string {
	strName := vFiled.GetName()
	return strings.ToUpper(strName[:1]) + strName[1:]
}

func (s *Common) GetHostTableName() string {
	if s.hostModelTableName == "" {
		s.hostModelTableName = s.GetModelTableName(s.hostModel)
	}

	return s.hostModelTableName
}

func (s *Common) GetHostPrimaryKeyField() model.Field {
	return s.hostModel.GetPrimaryField()
}

func (s *Common) GetHostFields() model.Fields {
	return s.hostModel.GetFields()
}

func (s *Common) GetHostModelValue() (ret string, err *cd.Result) {
	if s.hostModelValue == "" {
		entityVal, entityErr := s.buildModelValue(s.hostModel)
		if entityErr != nil {
			err = entityErr
			return
		}

		s.hostModelValue = entityVal
	}

	ret = s.hostModelValue
	return
}

func (s *Common) GetModelTableName(vModel model.Model) string {
	tableName := s.constructTableName(vModel)
	if s.specialPrefix != "" {
		tableName = fmt.Sprintf("%s_%s", s.specialPrefix, tableName)
	}

	return tableName
}

func (s *Common) GetRelationTableName(vField model.Field, rModel model.Model) string {
	leftName := s.constructTableName(s.hostModel)
	rightName := s.constructTableName(rModel)
	infixVal := s.constructInfix(vField)

	tableName := fmt.Sprintf("%s%s%s%s", leftName, infixVal, getFieldRelation(vField), rightName)
	if s.specialPrefix != "" {
		tableName = fmt.Sprintf("%s_%s", s.specialPrefix, tableName)
	}

	return tableName
}

func (s *Common) GetRelationValue(rModel model.Model) (leftVal, rightVal string, err *cd.Result) {
	entityVal, entityErr := s.GetHostModelValue()
	if entityErr != nil {
		err = entityErr
		log.Errorf("GetRelationValue failed, s.GetHostModelValue error:%s", err.Error())
		return
	}

	relationVal, relationErr := s.buildModelValue(rModel)
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

func (s *Common) buildModelValue(vModel model.Model) (ret string, err *cd.Result) {
	pkField := vModel.GetPrimaryField()
	switch pkField.GetType().GetValue() {
	case model.TypeStringValue:
		ret, err = s.encodeStringValue(pkField.GetType(), pkField.GetValue())
	case model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeBigIntegerValue, model.TypeIntegerValue,
		model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveBigIntegerValue, model.TypePositiveIntegerValue:
		ret, err = s.encodeIntValue(pkField.GetType(), pkField.GetValue())
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal pkFiled type %s", pkField.GetType().GetPkgKey()))
	}
	return
}

func (s *Common) BuildFieldValue(vType model.Type, vValue model.Value) (ret string, err *cd.Result) {
	if !vValue.IsValid() {
		ret, err = getTypeDefaultValue(vType)
		return
	}

	switch vType.GetValue() {
	case model.TypeStringValue, model.TypeDateTimeValue:
		ret, err = s.encodeStringValue(vType, vValue)
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
	if !vValue.IsValid() {
		err = cd.NewError(cd.UnExpected, "nil opr value")
		return
	}

	switch vType.GetValue() {
	case model.TypeStringValue, model.TypeDateTimeValue:
		ret, err = s.encodeStringValue(vType, vValue)
	case model.TypeBooleanValue, model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeBigIntegerValue, model.TypeIntegerValue,
		model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveBigIntegerValue, model.TypePositiveIntegerValue:
		ret, err = s.encodeIntValue(vType, vValue)
	case model.TypeFloatValue, model.TypeDoubleValue:
		ret, err = s.encodeFloatValue(vType, vValue)
	case model.TypeStructValue:
		ret, err = s.encodeStructValue(vType, vValue)
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

func (s *Common) BuildModelFilter() (ret string, err *cd.Result) {
	pkField := s.hostModel.GetPrimaryField()
	pkfVal, pkfErr := s.BuildFieldValue(pkField.GetType(), pkField.GetValue())
	if pkfErr != nil {
		err = pkfErr
		log.Errorf("BuildModelFilter failed, s.EncodeValue error:%s", err.Error())
		return
	}

	pkfName := pkField.GetName()
	ret = fmt.Sprintf("`%s` = %v", pkfName, pkfVal)
	return
}

func (s *Common) encodeStringValue(vType model.Type, vValue model.Value) (ret string, err *cd.Result) {
	fEncodeVal, fEncodeErr := s.EncodeValue(vType, vValue)
	if fEncodeErr != nil {
		err = fEncodeErr
		log.Errorf("encodeStringValue failed, s.EncodeValue error:%s", err.Error())
		return
	}

	ret = fmt.Sprintf("'%s'", fEncodeVal)
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

func (s *Common) encodeStructValue(vType model.Type, vValue model.Value) (ret string, err *cd.Result) {
	fEncodeVal, fEncodeErr := s.EncodeValue(vType, vValue)
	if fEncodeErr != nil {
		err = fEncodeErr
		log.Errorf("encodeStringValue failed, s.EncodeValue error:%s", err.Error())
		return
	}
	switch fEncodeVal.(type) {
	case string:
		ret = fmt.Sprintf("'%s'", fEncodeVal)
	default:
		ret = fmt.Sprintf("%v", fEncodeVal)
	}
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

func (s *Common) encodeSliceStruct(sliceVal []interface{}) []string {
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
	case model.TypeStructValue:
		ret = s.encodeSliceStruct(sliceVal)
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal filed type %s", vType.Elem().GetPkgKey()))
	}

	return
}
