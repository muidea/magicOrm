package codec

import (
	"encoding/json"
	"fmt"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

type Codec interface {
	BuildHostModelTableName() string
	BuildModelTableName(vModel model.Model) string
	BuildRelationTableName(vField model.Field, rModel model.Model) (ret string, err *cd.Result)
	BuildHostModelValue() (ret model.RawVal, err *cd.Result)
	BuildRelationValue(rModel model.Model) (leftVal, rightVal model.RawVal, err *cd.Result)
	BuildFieldValue(vField model.Field) (ret model.RawVal, err *cd.Result)
	BuildOprValue(vType model.Type, vValue model.Value) (ret model.RawVal, err *cd.Result)
	ExtractFiledValue(vType model.Type, eVal model.RawVal) (ret model.Value, err *cd.Result)
}

type codecImpl struct {
	hostModel     model.Model
	modelProvider provider.Provider
	specialPrefix string
}

func New(vModel model.Model, modelProvider provider.Provider, prefix string) Codec {
	return &codecImpl{hostModel: vModel, modelProvider: modelProvider, specialPrefix: prefix}
}

func (s *codecImpl) constructTableName(vModel model.Model) string {
	strName := vModel.GetName()
	return strings.ToUpper(strName[:1]) + strName[1:]
}

func (s *codecImpl) constructInfix(vFiled model.Field) string {
	strName := vFiled.GetName()
	return strings.ToUpper(strName[:1]) + strName[1:]
}

func (s *codecImpl) BuildHostModelTableName() string {
	return s.BuildModelTableName(s.hostModel)
}

func (s *codecImpl) BuildModelTableName(vModel model.Model) string {
	tableName := s.constructTableName(vModel)
	if s.specialPrefix != "" {
		tableName = fmt.Sprintf("%s_%s", s.specialPrefix, tableName)
	}

	return tableName
}

func (s *codecImpl) BuildRelationTableName(vField model.Field, rModel model.Model) (ret string, err *cd.Result) {
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

	tableName := fmt.Sprintf("%s%s%s%s", leftName, infixVal, s.getFieldRelation(vField), rightName)
	if s.specialPrefix != "" {
		tableName = fmt.Sprintf("%s_%s", s.specialPrefix, tableName)
	}

	ret = tableName
	return
}

func (s *codecImpl) BuildHostModelValue() (ret model.RawVal, err *cd.Result) {
	entityVal, entityErr := s.buildModelValue(s.hostModel)
	if entityErr != nil {
		err = entityErr
		return
	}

	ret = entityVal
	return
}

func (s *codecImpl) BuildRelationValue(rModel model.Model) (leftVal, rightVal model.RawVal, err *cd.Result) {
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

func (s *codecImpl) BuildFieldValue(vField model.Field) (ret model.RawVal, err *cd.Result) {
	vType := vField.GetType()
	vValue := vField.GetValue()
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

func (s *codecImpl) BuildOprValue(vType model.Type, vValue model.Value) (ret model.RawVal, err *cd.Result) {
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

func (s *codecImpl) ExtractFiledValue(vType model.Type, eVal model.RawVal) (ret model.Value, err *cd.Result) {
	return
}

func (s *codecImpl) buildModelValue(vModel model.Model) (ret model.RawVal, err *cd.Result) {
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

func (s *codecImpl) encodeStringValue(vType model.Type, vValue model.Value) (ret model.RawVal, err *cd.Result) {
	fEncodeVal, fEncodeErr := s.modelProvider.EncodeValue(vValue, vType)
	if fEncodeErr != nil {
		err = fEncodeErr
		log.Errorf("encodeStringValue failed, s.EncodeValue error:%s", err.Error())
		return
	}

	ret = fEncodeVal
	return
}

func (s *codecImpl) encodeIntValue(vType model.Type, vValue model.Value) (ret model.RawVal, err *cd.Result) {
	fEncodeVal, fEncodeErr := s.modelProvider.EncodeValue(vValue, vType)
	if fEncodeErr != nil {
		err = fEncodeErr
		log.Errorf("encodeIntValue failed, s.EncodeValue error:%s", err.Error())
		return
	}
	ret = fEncodeVal
	return
}

func (s *codecImpl) encodeFloatValue(vType model.Type, vValue model.Value) (ret model.RawVal, err *cd.Result) {
	fEncodeVal, fEncodeErr := s.modelProvider.EncodeValue(vValue, vType)
	if fEncodeErr != nil {
		err = fEncodeErr
		log.Errorf("encodeFloatValue failed, s.EncodeValue error:%s", err.Error())
		return
	}
	ret = fEncodeVal
	return
}

func (s *codecImpl) encodeStructValue(vType model.Type, vValue model.Value) (ret model.RawVal, err *cd.Result) {
	fEncodeVal, fEncodeErr := s.modelProvider.EncodeValue(vValue, vType)
	if fEncodeErr != nil {
		err = fEncodeErr
		log.Errorf("encodeStructValue failed, s.EncodeValue error:%s", err.Error())
		return
	}

	ret = fEncodeVal
	return
}

func (s *codecImpl) encodeSliceString(sliceVal []any) []string {
	strSlice := make([]string, len(sliceVal))
	for idx, val := range sliceVal {
		strSlice[idx] = val.(string)
	}

	return strSlice
}

func (s *codecImpl) encodeSliceInt(sliceVal []any) []string {
	strSlice := make([]string, len(sliceVal))
	for idx, val := range sliceVal {
		strSlice[idx] = fmt.Sprintf("%v", val)
	}

	return strSlice
}

func (s *codecImpl) encodeSliceFloat(sliceVal []any) []string {
	strSlice := make([]string, len(sliceVal))
	for idx, val := range sliceVal {
		strSlice[idx] = fmt.Sprintf("%v", val)
	}

	return strSlice
}

func (s *codecImpl) encodeSliceStruct(sliceVal []any) []string {
	strSlice := make([]string, len(sliceVal))
	for idx, val := range sliceVal {
		strSlice[idx] = fmt.Sprintf("%v", val)
	}

	return strSlice
}

func (s *codecImpl) encodeSliceValue(vType model.Type, vValue model.Value) (ret []string, err *cd.Result) {
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

func (s *codecImpl) getFieldRelation(vField model.Field) (ret relationType) {
	fType := vField.GetType()
	if fType.IsBasic() {
		return
	}

	isPtr := fType.Elem().IsPtrType() || fType.IsPtrType()
	isSlice := model.IsSliceType(fType.GetValue())

	if !isPtr && !isSlice {
		ret = relationHas1v1
		return
	}

	if !isPtr && isSlice {
		ret = relationHas1vn
		return
	}

	if isPtr && !isSlice {
		ret = relationRef1v1
		return
	}

	ret = relationRef1vn
	return
}
