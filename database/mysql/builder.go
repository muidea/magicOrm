package mysql

import (
	"fmt"
	"strings"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/util"
)

// Builder Builder
type Builder struct {
	modelInfo     model.Model
	modelProvider provider.Provider
}

// New create builder
func New(vModel model.Model, modelProvider provider.Provider) *Builder {
	return &Builder{modelInfo: vModel, modelProvider: modelProvider}
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

func (s *Builder) getRelationValue(rModel model.Model) (leftVal, rightVal interface{}, err error) {
	infoVal, infoErr := s.getModelValue(s.modelInfo)
	if infoErr != nil {
		err = infoErr
		return
	}
	relationVal, relationErr := s.getModelValue(rModel)
	if relationErr != nil {
		err = relationErr
		return
	}

	leftVal = infoVal
	rightVal = relationVal
	return
}

func (s *Builder) GetInitializeValue(field model.Field) (ret interface{}, err error) {
	return getFieldInitializeValue(field)
}

func (s *Builder) getModelValue(vModel model.Model) (ret interface{}, err error) {
	pkField := vModel.GetPrimaryField()
	if pkField == nil {
		err = fmt.Errorf("no define primaryKey")
		return
	}

	fStr, fErr := s.modelProvider.EncodeValue(pkField.GetValue(), pkField.GetType())
	if fErr != nil {
		err = fErr
		return
	}

	ret = fStr
	return
}

func (s *Builder) buildValue(vValue model.Value, vType model.Type) (ret interface{}, err error) {
	fStr, fErr := s.modelProvider.EncodeValue(vValue, vType)
	if fErr != nil {
		err = fErr
		return
	}

	switch vType.GetValue() {
	case util.TypeStringField, util.TypeDateTimeField, util.TypeSliceField:
		ret = fmt.Sprintf("'%v'", fStr)
	default:
		ret = fStr
	}

	return
}

func (s *Builder) buildPKFilter(vModel model.Model) (ret string, err error) {
	pkfVal, pkfErr := s.getModelValue(vModel)
	if pkfErr != nil {
		err = pkfErr
		return
	}

	pkfTag := vModel.GetPrimaryField().GetTag().GetName()
	ret = fmt.Sprintf("`%s`=%v", pkfTag, pkfVal)
	return
}
