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
		log.Errorf("BuildDelete failed, s.buildModelFilter error:%s", err.Error())
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
		log.Errorf("BuildDeleteRelation failed, s.GetModelValue error:%s", err.Error())
		return
	}

	relationTableName := s.GetRelationTableName(vField, rModel)
	delRight = fmt.Sprintf("DELETE FROM `%s` WHERE `%s` IN (SELECT `right` FROM `%s` WHERE `left`=%v)", s.GetHostTableName(rModel), s.GetPrimaryKeyField(rModel).GetName(), relationTableName, leftVal)
	//log.Print(delRight)

	delRelation = fmt.Sprintf("DELETE FROM `%s` WHERE `left`=%v", relationTableName, leftVal)
	//log.Print(delRelation)

	return
}
