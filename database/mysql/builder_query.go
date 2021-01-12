package mysql

import (
	"fmt"
	"log"

	"github.com/muidea/magicOrm/model"
)

// BuildQuery BuildQuery
func (s *Builder) BuildQuery(filter model.Filter) (ret string, err error) {
	namesVal, nameErr := s.getFieldQueryNames(s.modelInfo)
	if nameErr != nil {
		err = nameErr
		return
	}

	filterStr, filterErr := s.buildFilter(filter)
	if filterErr != nil {
		err = filterErr
		log.Printf("buildFilter failed, err:%s", err.Error())
		return
	}

	if filterStr != "" {
		ret = fmt.Sprintf("SELECT %s FROM `%s` WHERE %s", namesVal, s.getHostTableName(s.modelInfo), filterStr)
	} else {
		ret = fmt.Sprintf("SELECT %s FROM `%s`", namesVal, s.getHostTableName(s.modelInfo))
	}
	//log.Print(ret)

	return
}

// BuildQueryRelation BuildQueryRelation
func (s *Builder) BuildQueryRelation(fieldName string, relationInfo model.Model) (ret string, err error) {
	pkfStr, pkfErr := s.getModelStr(s.modelInfo)
	if pkfErr != nil {
		err = pkfErr
		return
	}

	ret = fmt.Sprintf("SELECT `right` FROM `%s` WHERE `left`= %s", s.GetRelationTableName(fieldName, relationInfo), pkfStr)
	//log.Print(ret)

	return
}

func (s *Builder) getFieldQueryNames(info model.Model) (ret string, err error) {
	str := ""
	for _, field := range info.GetFields() {
		fType := field.GetType()
		if !fType.IsBasic() {
			continue
		}

		fTag := field.GetTag()
		if str == "" {
			str = fmt.Sprintf("`%s`", fTag.GetName())
		} else {
			str = fmt.Sprintf("%s,`%s`", str, fTag.GetName())
		}
	}

	ret = str

	return
}
