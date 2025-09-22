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

type Identifier interface {
	GetName() string
	GetPkgPath() string
	GetDescription() string
}

type Codec interface {
	ConstructModelTableName(vIdentifier Identifier) string
	ConstructRelationTableName(vModel model.Model, vField model.Field) (string, *cd.Error)

	PackedBasicFieldValue(vField model.Field, vVal model.Value) (any, *cd.Error)
	PackedStructFieldValue(vField model.Field, vVal model.Value) (any, *cd.Error)
	PackedSliceStructFieldValue(vField model.Field, vVal model.Value) (any, *cd.Error)
	ExtractBasicFieldValue(vField model.Field, eVal any) (any, *cd.Error)
}

type codecImpl struct {
	specialPrefix string
	modelProvider provider.Provider
}

func New(provider provider.Provider, prefix string) Codec {
	return &codecImpl{
		modelProvider: provider,
		specialPrefix: prefix,
	}
}

func (s *codecImpl) constructTableName(vIdentifier Identifier) string {
	strName := vIdentifier.GetName()
	return strings.ToUpper(strName[:1]) + strName[1:]
}

func (s *codecImpl) constructInfix(vField model.Field) string {
	strName := vField.GetName()
	return strings.ToUpper(strName[:1]) + strName[1:]
}

func (s *codecImpl) ConstructModelTableName(vIdentifier Identifier) string {
	tableName := s.constructTableName(vIdentifier)
	if s.specialPrefix != "" {
		tableName = fmt.Sprintf("%s_%s", s.specialPrefix, tableName)
	}

	return tableName
}

func (s *codecImpl) ConstructRelationTableName(vModel model.Model, vField model.Field) (ret string, err *cd.Error) {
	leftName := s.constructTableName(vModel)
	rightName := s.constructTableName(vField.GetType())
	infixVal := s.constructInfix(vField)

	tableName := fmt.Sprintf("%s%s%s%s", leftName, infixVal, s.getFieldRelation(vField), rightName)
	if s.specialPrefix != "" {
		tableName = fmt.Sprintf("%s_%s", s.specialPrefix, tableName)
	}

	ret = tableName
	return
}

func (s *codecImpl) getFieldRelation(vField model.Field) (ret relationType) {
	fType := vField.GetType()
	isPtr := fType.Elem().IsPtrType()
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

func (s *codecImpl) PackedBasicFieldValue(vField model.Field, fVal model.Value) (ret any, err *cd.Error) {
	if !model.IsBasicField(vField) {
		err = cd.NewError(cd.Unexpected, "illegal field type")
		log.Errorf("PackedFieldValue failed, error:%s", err.Error())
		return
	}

	switch vField.GetType().GetValue() {
	case model.TypeSliceValue:
		sliceVal, sliceErr := s.modelProvider.EncodeValue(fVal.Get(), vField.GetType())
		if sliceErr != nil {
			err = sliceErr
			return
		}
		jsonVal, jsonErr := json.Marshal(sliceVal)
		if jsonErr != nil {
			err = cd.NewError(cd.Unexpected, jsonErr.Error())
			return
		}

		ret = string(jsonVal)
	default:
		ret, err = s.modelProvider.EncodeValue(fVal.Get(), vField.GetType())
	}

	if err != nil {
		log.Errorf("PackedFieldValue failed, error:%s", err.Error())
	}

	return
}

func (s *codecImpl) PackedStructFieldValue(vField model.Field, fVal model.Value) (ret any, err *cd.Error) {
	if !model.IsStructField(vField) || !model.IsStruct(vField.GetType().Elem()) {
		err = cd.NewError(cd.Unexpected, "illegal field type")
		log.Errorf("PackedStructFieldValue failed, error:%s", err.Error())
		return
	}

	vModelVal, modelValErr := s.modelProvider.GetTypeModel(vField.GetType().Elem())
	if modelValErr != nil {
		err = modelValErr
		log.Errorf("PackedStructFieldValue failed, s.modelProvider.GetTypeModel error:%s", err.Error())
		return
	}

	vModelVal, modelValErr = s.modelProvider.SetModelValue(vModelVal, fVal)
	if modelValErr != nil {
		err = modelValErr
		log.Errorf("PackedStructFieldValue failed, s.modelProvider.SetModelValue error:%s", err.Error())
		return
	}

	pkField := vModelVal.GetPrimaryField()
	ret, err = s.PackedBasicFieldValue(pkField, pkField.GetValue())
	return
}

func (s *codecImpl) PackedSliceStructFieldValue(vField model.Field, fVal model.Value) (ret any, err *cd.Error) {
	if !model.IsSliceField(vField) || model.IsStruct(vField.GetType().Elem()) {
		err = cd.NewError(cd.Unexpected, "illegal field type")
		log.Errorf("PackedSliceStructFieldValue failed, error:%s", err.Error())
		return
	}

	vModelVal, modelValErr := s.modelProvider.GetTypeModel(vField.GetType().Elem())
	if modelValErr != nil {
		err = modelValErr
		log.Errorf("PackedSliceStructFieldValue failed, s.modelProvider.GetTypeModel error:%s", err.Error())
		return
	}

	valueList := []any{}
	sliceVal := vField.GetSliceValue()
	for _, val := range sliceVal {
		vModelVal, modelValErr = s.modelProvider.SetModelValue(vModelVal.Copy(model.LiteView), val)
		if modelValErr != nil {
			err = modelValErr
			log.Errorf("PackedSliceStructFieldValue failed, s.modelProvider.SetModelValue error:%s", err.Error())
			return
		}
		pkField := vModelVal.GetPrimaryField()
		pkVal, pkErr := s.PackedBasicFieldValue(pkField, pkField.GetValue())
		if pkErr != nil {
			err = pkErr
			log.Errorf("PackedSliceStructFieldValue failed, s.PackedBasicFieldValue error:%s", err.Error())
			return
		}
		valueList = append(valueList, pkVal)
	}

	ret = valueList
	return
}

func (s *codecImpl) ExtractBasicFieldValue(vField model.Field, eVal any) (ret any, err *cd.Error) {
	vType := vField.GetType()
	switch vType.GetValue() {
	case model.TypeSliceValue:
		strVal, strOK := eVal.(*string)
		if strOK {
			if *strVal != "" {
				vArray := []any{}
				byteErr := json.Unmarshal([]byte(*strVal), &vArray)
				if byteErr != nil {
					err = cd.NewError(cd.Unexpected, byteErr.Error())
					return
				}
				ret, err = s.modelProvider.DecodeValue(vArray, vType)
			} else {
				ret, err = vType.Interface(nil)
			}
		} else {
			err = cd.NewError(cd.Unexpected, "illegal field value")
		}
	default:
		ret, err = s.modelProvider.DecodeValue(eVal, vType)
	}
	return
}
