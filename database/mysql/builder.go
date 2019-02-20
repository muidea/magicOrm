package mysql

import (
	"fmt"
	"strings"

	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/provider"
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
	return strings.Join(strings.Split(info.GetName(), "."), "_")
}

// GetTableName GetTableName
func (s *Builder) GetTableName() string {
	return s.getTableName(s.modelInfo)
}

// GetRelationTableName GetRelationTableName
func (s *Builder) GetRelationTableName(fieldName string, relationInfo model.Model) string {
	leftName := s.getTableName(s.modelInfo)
	rightName := s.getTableName(relationInfo)

	return fmt.Sprintf("%s%s2%s", leftName, fieldName, rightName)
}

func (s *Builder) getStructValue(modelInfo model.Model) (ret string, err error) {
	structKey := modelInfo.GetPrimaryField()
	if structKey == nil {
		err = fmt.Errorf("no define primaryKey")
		return
	}

	fType := structKey.GetType()
	fValue := structKey.GetValue()
	fStr, fErr := s.modelProvider.GetValueStr(fType, fValue)
	if fErr != nil {
		err = fErr
		return
	}

	ret = fStr
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
