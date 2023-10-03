package mysql

import (
	"fmt"

	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BuildDelete  BuildDelete
func (s *Builder) BuildDelete() (ret string, err error) {
	filterStr, filterErr := s.buildModelFilter()
	if filterErr != nil {
		err = filterErr
		log.Errorf("buildModelFilter failed, err:%s", err.Error())
		return
	}

	str := fmt.Sprintf("DELETE FROM `%s` WHERE %s", s.GetTableName(), filterStr)
	//log.Print(str)

	ret = str
	return
}

// BuildDeleteRelation BuildDeleteRelation
func (s *Builder) BuildDeleteRelation(vField model.Field, rModel model.Model) (delRight, delRelation string, err error) {
	leftVal, leftErr := s.GetModelValue()
	if leftErr != nil {
		err = leftErr
		return
	}
	relationTableName := s.GetRelationTableName(vField, rModel)
	delRight = fmt.Sprintf("DELETE FROM `%s` WHERE `id` IN (SELECT `right` FROM `%s` WHERE `left`=%v)", s.GetHostTableName(rModel), relationTableName, leftVal)
	//log.Print(delRight)

	delRelation = fmt.Sprintf("DELETE FROM `%s` WHERE `left`=%v", relationTableName, leftVal)
	//log.Print(delRelation)

	return
}
