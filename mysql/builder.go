package mysql

import (
	"fmt"
	"strings"

	"muidea.com/magicOrm/model"
)

// Builder Builder
type Builder struct {
	structInfo model.Model
}

// New create builder
func New(structInfo model.Model) *Builder {
	//err := verifyStructInfo(structInfo)
	//if err != nil {
	//	log.Printf("verify structInfo failed, err:%s", err.Error())
	//	return nil
	//}

	return &Builder{structInfo: structInfo}
}

func (s *Builder) getTableName(info model.Model) string {
	return strings.Join(strings.Split(info.GetName(), "."), "_")
}

// GetTableName GetTableName
func (s *Builder) GetTableName() string {
	return s.getTableName(s.structInfo)
}

// GetRelationTableName GetRelationTableName
func (s *Builder) GetRelationTableName(fieldName string, relationInfo model.Model) string {
	leftName := s.getTableName(s.structInfo)
	rightName := s.getTableName(relationInfo)

	return fmt.Sprintf("%s%s2%s", leftName, fieldName, rightName)
}

func (s *Builder) getStructValue(structInfo model.Model) (ret string, err error) {
	structKey := structInfo.GetPrimaryField()
	if structKey == nil {
		err = fmt.Errorf("no define primaryKey")
		return
	}

	fValue := structKey.GetFieldValue()
	if fValue == nil {
		err = fmt.Errorf("nil primaryKey value")
		return
	}

	structVal, structErr := fValue.GetValueStr()
	if structErr != nil {
		err = structErr
		return
	}

	ret = structVal
	return
}

func (s *Builder) getRelationValue(relationInfo model.Model) (leftVal, rightVal string, err error) {
	structKey := s.structInfo.GetPrimaryField()
	relationKey := relationInfo.GetPrimaryField()
	if structKey == nil || relationKey == nil {
		err = fmt.Errorf("no define primaryKey")
		return
	}

	structVal, structErr := structKey.GetFieldValue().GetValueStr()
	if structErr != nil {
		err = structErr
		return
	}
	relationVal, relationErr := relationKey.GetFieldValue().GetValueStr()
	if relationErr != nil {
		err = relationErr
		return
	}

	leftVal = structVal
	rightVal = relationVal
	return
}
