package mysql

import (
	"fmt"

	log "github.com/cihub/seelog"

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

	ret = fmt.Sprintf("DELETE FROM `%s` WHERE %s", s.GetTableName(), filterStr)
	//log.Print(ret)

	return
}

// BuildDeleteRelation BuildDeleteRelation
func (s *Builder) BuildDeleteRelation(vField model.Field, rModel model.Model) (delRight, delRelation string, err error) {
	leftVal, leftErr := s.GetModelValue()
	if leftErr != nil {
		err = leftErr
		return
	}
	relationSchema := s.GetRelationTableName(vField, rModel)
	delRight = fmt.Sprintf("DELETE FROM `%s` WHERE `id` in (SELECT `right` FROM `%s` WHERE `left`=%v)", s.GetHostTableName(rModel), relationSchema, leftVal)
	//log.Print(delRight)

	delRelation = fmt.Sprintf("DELETE FROM `%s` WHERE `left`=%v", relationSchema, leftVal)
	//log.Print(delRelation)

	return
}
