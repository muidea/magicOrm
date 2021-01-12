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

func (s *Builder) constructTableName(info model.Model) string {
	items := strings.Split(info.GetName(), ".")
	return items[len(items)-1]
}

// GetTableName GetTableName
func (s *Builder) GetTableName() string {
	return s.getHostTableName(s.modelInfo)
}

// getHostTableName getHostTableName
func (s *Builder) getHostTableName(info model.Model) string {
	tableName := s.constructTableName(info)
	return fmt.Sprintf("%s_%s", s.modelProvider.Owner(), tableName)
}

// GetRelationTableName GetRelationTableName
func (s *Builder) GetRelationTableName(fieldName string, relationInfo model.Model) string {
	leftName := s.constructTableName(s.modelInfo)
	rightName := s.constructTableName(relationInfo)

	return fmt.Sprintf("%s_%s%s2%s", s.modelProvider.Owner(), leftName, fieldName, rightName)
}

func (s *Builder) getFieldValue(fField model.Field) (ret string, err error) {
	fType := fField.GetType()
	fValue := fField.GetValue()
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
	structVal, structErr := s.getModelStr(s.modelInfo)
	if structErr != nil {
		err = structErr
		return
	}
	relationVal, relationErr := s.getModelStr(relationInfo)
	if relationErr != nil {
		err = relationErr
		return
	}

	leftVal = structVal
	rightVal = relationVal
	return
}

func (s *Builder) GetInitializeValue(field model.Field) (ret interface{}, err error) {
	return getFieldInitializeValue(field)
}

func (s *Builder) buildPKFilter() (ret string, err error) {
	pkfVal, pkfErr := s.getModelStr(s.modelInfo)
	if pkfErr != nil {
		err = pkfErr
		return
	}

	pkfTag := s.modelInfo.GetPrimaryField().GetTag().GetName()
	ret = fmt.Sprintf("`%s`=%s", pkfTag, pkfVal)
	return
}

func (s *Builder) getModelStr(vModel model.Model) (ret string, err error) {
	pkField := vModel.GetPrimaryField()
	if pkField == nil {
		err = fmt.Errorf("no define primaryKey")
		return
	}

	fStr, fErr := s.modelProvider.GetValueStr(pkField.GetValue(), pkField.GetType())
	if fErr != nil {
		err = fErr
		return
	}

	ret = fStr
	return
}
