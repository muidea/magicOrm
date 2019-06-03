package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// BuildQuery BuildQuery
func (s *Builder) BuildQuery() (ret string, err error) {
	namesVal, nameErr := s.getFieldQueryNames(s.modelInfo)
	if nameErr != nil {
		err = nameErr
		return
	}

	filterStr, filterErr := s.buildFilter(s.modelInfo)
	if filterErr != nil {
		err = filterErr
		return
	}

	ret = fmt.Sprintf("SELECT %s FROM `%s` WHERE %s", namesVal, s.getTableName(s.modelInfo), filterStr)
	//log.Print(ret)

	return
}

func (s *Builder) buildFilter(modelInfo model.Model) (ret string, err error) {
	filterSQL := ""
	for _, field := range modelInfo.GetFields() {
		if !field.IsAssigned() {
			continue
		}

		fType := field.GetType()
		fValue := field.GetValue()
		dependModel, dependErr := s.modelProvider.GetTypeModel(fType)
		if dependErr != nil {
			err = dependErr
			return
		}

		if dependModel != nil {
			continue
		}

		fStr, fErr := s.modelProvider.GetValueStr(fType, fValue)
		if fErr != nil {
			err = fErr
			return
		}

		fTag := field.GetTag()
		if filterSQL == "" {
			filterSQL = fmt.Sprintf("`%s`=%s", fTag.GetName(), fStr)
		} else {
			filterSQL = fmt.Sprintf("%s and `%s`=%s", filterSQL, fTag.GetName(), fStr)
		}
	}

	ret = filterSQL
	return
}

// BuildQueryRelation BuildQueryRelation
func (s *Builder) BuildQueryRelation(fieldName string, relationInfo model.Model) (ret string, err error) {
	pkfStr, pkfErr := s.getStructValue(s.modelInfo)
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
	for _, field := range s.modelInfo.GetFields() {
		fType := field.GetType()
		dependModel, dependErr := s.modelProvider.GetTypeModel(fType)
		if dependErr != nil {
			err = dependErr
			return
		}

		if dependModel != nil {
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
