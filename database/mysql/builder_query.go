package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicOrm/model"
)

// BuildQuery BuildQuery
func (s *Builder) BuildQuery() (ret string, err error) {
	pkfTag := s.modelInfo.GetPrimaryField().GetTag()
	pkfVal, pkfErr := s.getStructValue(s.modelInfo)
	if pkfErr != nil {
		err = pkfErr
		return
	}

	namesVal, nameErr := s.getFieldQueryNames(s.modelInfo)
	if nameErr != nil {
		err = nameErr
		return
	}

	ret = fmt.Sprintf("SELECT %s FROM `%s` WHERE `%s`=%s", namesVal, s.getTableName(s.modelInfo), pkfTag.GetName(), pkfVal)
	log.Print(ret)

	return
}

// BuildQueryRelation BuildQueryRelation
func (s *Builder) BuildQueryRelation(fieldName string, relationInfo model.Model) (ret string, err error) {
	pkfVal, pkfErr := s.getStructValue(s.modelInfo)
	if pkfErr != nil {
		err = pkfErr
		return
	}

	ret = fmt.Sprintf("SELECT `right` FROM `%s` WHERE `left`= %s", s.GetRelationTableName(fieldName, relationInfo), pkfVal)
	log.Print(ret)

	return
}

func (s *Builder) getFieldQueryNames(info model.Model) (ret string, err error) {
	str := ""
	for _, field := range s.modelInfo.GetFields() {
		fTag := field.GetTag()
		fType := field.GetType()

		dependModel, dependErr := s.modelProvider.GetTypeModel(fType.GetType())
		if dependErr != nil {
			err = dependErr
			return
		}

		if dependModel != nil {
			continue
		}

		if str == "" {
			str = fmt.Sprintf("`%s`", fTag.GetName())
		} else {
			str = fmt.Sprintf("%s,`%s`", str, fTag.GetName())
		}
	}

	ret = str

	return
}
