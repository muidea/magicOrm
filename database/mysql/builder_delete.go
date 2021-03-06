package mysql

import (
	"fmt"
	"log"

	"github.com/muidea/magicOrm/model"
)

// BuildDelete  BuildDelete
func (s *Builder) BuildDelete() (ret string, err error) {
	filterStr, filterErr := s.buildPKFilter(s.modelInfo)
	if filterErr != nil {
		err = filterErr
		log.Printf("buildPKFilter failed, err:%s", err.Error())
		return
	}

	ret = fmt.Sprintf("DELETE FROM `%s` WHERE %s", s.getHostTableName(s.modelInfo), filterStr)
	//log.Print(ret)

	return
}

// BuildDeleteRelation BuildDeleteRelation
func (s *Builder) BuildDeleteRelation(fieldName string, relationInfo model.Model) (delRight, delRelation string, err error) {
	leftVal, leftErr := s.getModelValue(s.modelInfo)
	if leftErr != nil {
		err = leftErr
		return
	}

	delRight = fmt.Sprintf("DELETE FROM `%s` WHERE `id` in (SELECT `right` FROM `%s` WHERE `left`=%v)", s.getHostTableName(relationInfo), s.GetRelationTableName(fieldName, relationInfo), leftVal)
	//log.Print(delRight)

	delRelation = fmt.Sprintf("DELETE FROM `%s` WHERE `left`=%v", s.GetRelationTableName(fieldName, relationInfo), leftVal)
	//log.Print(delRelation)

	return
}
