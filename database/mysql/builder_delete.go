package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicOrm/model"
)

// BuildDelete  BuildDelete
func (s *Builder) BuildDelete() (ret string, err error) {
	pkfValue := s.modelInfo.GetPrimaryField().GetValue()
	pkfTag := s.modelInfo.GetPrimaryField().GetTag()
	pkfStr, pkferr := pkfValue.GetValueStr()
	if pkferr == nil {
		ret = fmt.Sprintf("DELETE FROM `%s` WHERE `%s`=%s", s.getTableName(s.modelInfo), pkfTag.GetName(), pkfStr)
		log.Print(ret)
	}

	err = pkferr

	return
}

// BuildDeleteRelation BuildDeleteRelation
func (s *Builder) BuildDeleteRelation(fieldName string, relationInfo model.Model) (delRight, delRelation string, err error) {
	leftVal, leftErr := s.getStructValue(s.modelInfo)
	if leftErr != nil {
		err = leftErr
		return
	}

	delRight = fmt.Sprintf("DELETE FROM `%s` WHERE `id` in (SELECT `right` FROM `%s` WHERE `left`=%s)", s.getTableName(relationInfo), s.GetRelationTableName(fieldName, relationInfo), leftVal)
	log.Print(delRight)

	delRelation = fmt.Sprintf("DELETE FROM `%s` WHERE `left`=%s", s.GetRelationTableName(fieldName, relationInfo), leftVal)
	log.Print(delRelation)

	return
}
