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

	tableName := s.getTableName(s.modelInfo)

	ret = fmt.Sprintf("SELECT %s FROM `%s` WHERE %s", namesVal, tableName, filterStr)
	//log.Print(ret)

	return
}

func (s *Builder) buildFilter(modelInfo model.Model) (ret string, err error) {
	filterSQL := ""
	relationSQL := ""
	pkField := modelInfo.GetPrimaryField()
	for _, field := range modelInfo.GetFields() {
		if !field.IsAssigned() {
			continue
		}

		fType := field.GetType()
		fValue := field.GetValue()
		fStr, fErr := s.modelProvider.GetValueStr(fType, fValue)
		if fErr != nil {
			err = fErr
			return
		}

		dependModel, dependErr := s.modelProvider.GetTypeModel(fType)
		if dependErr != nil {
			err = dependErr
			return
		}

		if dependModel != nil {
			relationTable := s.GetRelationTableName(field.GetName(), dependModel)
			if relationSQL == "" {
				relationSQL = fmt.Sprintf("`%s`=(SELECT `left` FROM `%s` WHERE `right`=%s)", pkField.GetName(), relationTable, fStr)
			} else {
				relationSQL = fmt.Sprintf("%s AND `%s`=(SELECT `left` FROM `%s` WHERE `right`=%s)", relationSQL, pkField.GetName(), relationTable, fStr)
			}

			continue
		}

		fTag := field.GetTag()
		if filterSQL == "" {
			filterSQL = fmt.Sprintf("`%s`=%s", fTag.GetName(), fStr)
		} else {
			filterSQL = fmt.Sprintf("%s AND `%s`=%s", filterSQL, fTag.GetName(), fStr)
		}
	}

	if relationSQL != "" {
		filterSQL = fmt.Sprintf("%s AND %s", filterSQL, relationSQL)
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
