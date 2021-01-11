package mysql

import (
	"fmt"
	"github.com/muidea/magicOrm/util"
	"strings"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

// Builder Builder
type Builder struct {
	modelInfo     model.Model
	modelProvider provider.Provider
}

// New create builder
func New(modelInfo model.Model, modelProvider provider.Provider) *Builder {
	return &Builder{modelInfo: modelInfo, modelProvider: modelProvider}
}

func (s *Builder) getTableName(info model.Model) string {
	items := strings.Split(info.GetName(), ".")
	return items[len(items)-1]
}

// GetTableName GetTableName
func (s *Builder) GetTableName() string {
	return s.getHostTableName(s.modelInfo)
}

// getHostTableName getHostTableName
func (s *Builder) getHostTableName(info model.Model) string {
	tableName := s.getTableName(info)
	return fmt.Sprintf("%s_%s", s.modelProvider.Owner(), tableName)
}

// GetRelationTableName GetRelationTableName
func (s *Builder) GetRelationTableName(fieldName string, relationInfo model.Model) string {
	leftName := s.getTableName(s.modelInfo)
	rightName := s.getTableName(relationInfo)

	return fmt.Sprintf("%s_%s%s2%s", s.modelProvider.Owner(), leftName, fieldName, rightName)
}

func (s *Builder) getStructValue(modelInfo model.Model) (ret string, err error) {
	pkField := modelInfo.GetPrimaryField()
	if pkField == nil {
		err = fmt.Errorf("no define primaryKey")
		return
	}

	fStr, isNil, fErr := s.getFieldValue(pkField)
	if fErr != nil {
		err = fErr
		return
	}
	if isNil {
		err = fmt.Errorf("illegal primarykey value")
		return
	}

	ret = fStr
	return
}

func (s *Builder) getFieldValue(field model.Field) (ret string, isNil bool, err error) {
	fType := field.GetType()
	fValue := field.GetValue()

	if !fType.IsPtrType() {
		if fValue == nil || fValue.IsNil() {
			err = fmt.Errorf("illegal field value, must assigned first, name:%s", field.GetName())
			return
		}
	}

	if fType.IsPtrType() {
		if fValue == nil || fValue.IsNil() {
			isNil = true
			return
		}
	}

	dependModel, dependErr := s.modelProvider.GetTypeModel(fType)
	if dependErr != nil {
		err = dependErr
		return
	}
	if dependModel != nil {
		isNil = true
		return
	}

	fStr, fErr := s.modelProvider.GetValueStr(fValue, fType)
	if fErr != nil {
		err = fErr
		return
	}

	switch fType.GetValue() {
	case util.TypeStringField, util.TypeDateTimeField, util.TypeSliceField:
		ret = fmt.Sprintf("'%s'", fStr)
	default:
		ret = fStr
	}

	return
}

func (s *Builder) getRelationValue(relationInfo model.Model) (leftVal, rightVal string, err error) {
	structVal, structErr := s.getStructValue(s.modelInfo)
	if structErr != nil {
		err = structErr
		return
	}
	relationVal, relationErr := s.getStructValue(relationInfo)
	if relationErr != nil {
		err = relationErr
		return
	}

	leftVal = structVal
	rightVal = relationVal
	return
}

func (s *Builder) DeclareFieldValue(field model.Field) (ret interface{}, err error) {
	return getFieldInitValue(field)
}

func (s *Builder) buildPKFilter() (ret string, err error) {
	pkfVal, pkfErr := s.getStructValue(s.modelInfo)
	if pkfErr != nil {
		err = pkfErr
		return
	}

	pkfTag := s.modelInfo.GetPrimaryField().GetTag().GetName()
	ret = fmt.Sprintf("`%s`=%s", pkfTag, pkfVal)
	return
}
