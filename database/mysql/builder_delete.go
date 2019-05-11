package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// BuildDelete  BuildDelete
func (s *Builder) BuildDelete() (ret string, err error) {
	pkfVal, pkfErr := s.getStructValue(s.modelInfo)
	if pkfErr != nil {
		err = pkfErr
		return
	}

	pkfTag := s.modelInfo.GetPrimaryField().GetTag()
	ret = fmt.Sprintf("DELETE FROM `%s` WHERE `%s`=%s", s.getTableName(s.modelInfo), pkfTag.GetName(), pkfVal)
	//log.Print(ret)

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
	//log.Print(delRight)

	delRelation = fmt.Sprintf("DELETE FROM `%s` WHERE `left`=%s", s.GetRelationTableName(fieldName, relationInfo), leftVal)
	//log.Print(delRelation)

	return
}
