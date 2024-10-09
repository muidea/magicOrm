package context

import (
	"encoding/json"
	"fmt"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

type Context interface {
	BuildHostModelTableName() string
	BuildModelTableName(vModel model.Model) string
	BuildRelationTableName(vField model.Field, rModel model.Model) (ret string, err *cd.Result)
	BuildHostModelValue() (ret model.RawVal, err *cd.Result)
	BuildRelationValue(rModel model.Model) (leftVal, rightVal model.RawVal, err *cd.Result)
	BuildFieldValue(vType model.Type, vValue model.Value) (ret model.RawVal, err *cd.Result)
	BuildOprValue(vType model.Type, vValue model.Value) (ret model.RawVal, err *cd.Result)
	ExtractFiledValue(vType model.Type, eVal model.RawVal) (ret model.Value, err *cd.Result)
}

type contextImpl struct {
	hostModel     model.Model
	modelProvider provider.Provider
	specialPrefix string

	// temp value, for performance optimization
	hostModelTableName string
	hostModelValue     model.RawVal
}

func New(vModel model.Model, modelProvider provider.Provider, prefix string) Context {
	return &contextImpl{hostModel: vModel, modelProvider: modelProvider, specialPrefix: prefix}
}

func (s *contextImpl) constructTableName(vModel model.Model) string {
	strName := vModel.GetName()
	return strings.ToUpper(strName[:1]) + strName[1:]
}

func (s *contextImpl) constructInfix(vFiled model.Field) string {
	strName := vFiled.GetName()
	return strings.ToUpper(strName[:1]) + strName[1:]
}

func (s *contextImpl) BuildHostModelTableName() string {
	if s.hostModelTableName == "" {
		s.hostModelTableName = s.BuildModelTableName(s.hostModel)
	}

	return s.hostModelTableName
}

func (s *contextImpl) BuildModelTableName(vModel model.Model) string {
	tableName := s.constructTableName(vModel)
	if s.specialPrefix != "" {
		tableName = fmt.Sprintf("%s_%s", s.specialPrefix, tableName)
	}

	return tableName
}

func (s *contextImpl) BuildRelationTableName(vField model.Field, rModel model.Model) (ret string, err *cd.Result) {
	if rModel == nil {
		fieldModel, fieldErr := s.modelProvider.GetTypeModel(vField.GetType())
		if fieldErr != nil {
			err = fieldErr
			log.Errorf("BuildRelationTableName failed, s.modelProvider.GetTypeModel error:%s", err.Error())
			return
		}
		rModel = fieldModel
	}

	leftName := s.constructTableName(s.hostModel)
	rightName := s.constructTableName(rModel)
	infixVal := s.constructInfix(vField)

	tableName := fmt.Sprintf("%s%s%s%s", leftName, infixVal, getFieldRelation(vField), rightName)
	if s.specialPrefix != "" {
		tableName = fmt.Sprintf("%s_%s", s.specialPrefix, tableName)
	}

	ret = tableName
	return
}

func (s *contextImpl) BuildHostModelValue() (ret model.RawVal, err *cd.Result) {
	if s.hostModelValue == nil {
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

func (s *contextImpl) BuildRelationValue(rModel model.Model) (leftVal, rightVal model.RawVal, err *cd.Result) {
	entityVal, entityErr := s.BuildHostModelValue()
	if entityErr != nil {
		err = entityErr
		log.Errorf("BuildRelationValue failed, s.BuildHostModelValue error:%s", err.Error())
		return
	}

	relationVal, relationErr := s.buildModelValue(rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildRelationValue failed, s.EncodeModelValue error:%s", err.Error())
		return
	}

	leftVal = entityVal
	rightVal = relationVal
	return
}

func (s *contextImpl) BuildFieldValue(vType model.Type, vValue model.Value) (ret model.RawVal, err *cd.Result) {
	if !vValue.IsValid() {
		defaultVal, defaultErr := getBasicTypeDefaultValue(vType)
		if defaultErr != nil {
			err = defaultErr
			return
		}

		ret = model.NewRawVal(defaultVal)
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
		fEncodeVal, fEncodeErr := s.modelProvider.EncodeValue(vValue, vType)
		if fEncodeErr != nil {
			err = fEncodeErr
			log.Errorf("encodeIntValue failed, s.EncodeValue error:%s", err.Error())
			return
		}
		byteVal, byteErr := json.Marshal(fEncodeVal.Value())
		if byteErr != nil {
			err = cd.NewError(cd.UnExpected, fmt.Sprintf("%s", byteErr.Error()))
			return
		}
		ret = model.NewRawVal(strings.ReplaceAll(string(byteVal), "'", "''"))
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal filed type %s", vType.GetPkgKey()))
	}
	return
}

func (s *contextImpl) BuildOprValue(vType model.Type, vValue model.Value) (ret model.RawVal, err *cd.Result) {
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
		strVal := strings.Join(sliceVal, ",")
		ret = model.NewRawVal(strVal)
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal filed type %s", vType.GetPkgKey()))
	}
	return
}

func (s *contextImpl) ExtractFiledValue(vType model.Type, eVal model.RawVal) (ret model.Value, err *cd.Result) {
	return
}

func (s *contextImpl) buildModelValue(vModel model.Model) (ret model.RawVal, err *cd.Result) {
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

func (s *contextImpl) encodeStringValue(vType model.Type, vValue model.Value) (ret model.RawVal, err *cd.Result) {
	fEncodeVal, fEncodeErr := s.modelProvider.EncodeValue(vValue, vType)
	if fEncodeErr != nil {
		err = fEncodeErr
		log.Errorf("encodeStringValue failed, s.EncodeValue error:%s", err.Error())
		return
	}

	ret = fEncodeVal
	return
}

func (s *contextImpl) encodeIntValue(vType model.Type, vValue model.Value) (ret model.RawVal, err *cd.Result) {
	fEncodeVal, fEncodeErr := s.modelProvider.EncodeValue(vValue, vType)
	if fEncodeErr != nil {
		err = fEncodeErr
		log.Errorf("encodeIntValue failed, s.EncodeValue error:%s", err.Error())
		return
	}
	ret = fEncodeVal
	return
}

func (s *contextImpl) encodeFloatValue(vType model.Type, vValue model.Value) (ret model.RawVal, err *cd.Result) {
	fEncodeVal, fEncodeErr := s.modelProvider.EncodeValue(vValue, vType)
	if fEncodeErr != nil {
		err = fEncodeErr
		log.Errorf("encodeFloatValue failed, s.EncodeValue error:%s", err.Error())
		return
	}
	ret = fEncodeVal
	return
}

func (s *contextImpl) encodeStructValue(vType model.Type, vValue model.Value) (ret model.RawVal, err *cd.Result) {
	fEncodeVal, fEncodeErr := s.modelProvider.EncodeValue(vValue, vType)
	if fEncodeErr != nil {
		err = fEncodeErr
		log.Errorf("encodeStructValue failed, s.EncodeValue error:%s", err.Error())
		return
	}

	ret = fEncodeVal
	return
}

func (s *contextImpl) encodeSliceString(sliceVal []any) []string {
	strSlice := make([]string, len(sliceVal))
	for idx, val := range sliceVal {
		strSlice[idx] = val.(string)
	}

	return strSlice
}

func (s *contextImpl) encodeSliceInt(sliceVal []any) []string {
	strSlice := make([]string, len(sliceVal))
	for idx, val := range sliceVal {
		strSlice[idx] = fmt.Sprintf("%v", val)
	}

	return strSlice
}

func (s *contextImpl) encodeSliceFloat(sliceVal []any) []string {
	strSlice := make([]string, len(sliceVal))
	for idx, val := range sliceVal {
		strSlice[idx] = fmt.Sprintf("%v", val)
	}

	return strSlice
}

func (s *contextImpl) encodeSliceStruct(sliceVal []any) []string {
	strSlice := make([]string, len(sliceVal))
	for idx, val := range sliceVal {
		strSlice[idx] = fmt.Sprintf("%v", val)
	}

	return strSlice
}

func (s *contextImpl) encodeSliceValue(vType model.Type, vValue model.Value) (ret []string, err *cd.Result) {
	fEncodeVal, fEncodeErr := s.modelProvider.EncodeValue(vValue, vType)
	if fEncodeErr != nil {
		err = fEncodeErr
		log.Errorf("encodeSliceValue failed, s.EncodeValue error:%s", err.Error())
		return
	}

	sliceVal, sliceOK := fEncodeVal.Value().([]any)
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