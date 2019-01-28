package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicOrm/model"
)

// BuildQuery BuildQuery
func (s *Builder) BuildQuery() (ret string, err error) {
	pk := s.modelInfo.GetPrimaryField()
	if pk == nil {
		err = fmt.Errorf("no define primaryKey")
		return
	}

	pkfValue := pk.GetFieldValue()
	pkfTag := pk.GetFieldTag()
	pkfStr, pkferr := pkfValue.GetValueStr()
	if pkferr == nil {
		ret = fmt.Sprintf("SELECT %s FROM `%s` WHERE `%s`=%s", s.getFieldQueryNames(s.modelInfo), s.getTableName(s.modelInfo), pkfTag.Name(), pkfStr)
		log.Print(ret)
	}
	err = pkferr

	return
}

// BuildQueryRelation BuildQueryRelation
func (s *Builder) BuildQueryRelation(fieldName string, relationInfo model.Model) (ret string, err error) {
	pk := s.modelInfo.GetPrimaryField()
	if pk == nil {
		err = fmt.Errorf("no define primaryKey")
		return
	}

	pkfValue := pk.GetFieldValue()
	pkfStr, pkferr := pkfValue.GetValueStr()
	if pkferr == nil {
		ret = fmt.Sprintf("SELECT `right` FROM `%s` WHERE `left`= %s", s.GetRelationTableName(fieldName, relationInfo), pkfStr)
		log.Print(ret)
	}

	err = pkferr

	return
}

func (s *Builder) getFieldQueryNames(info model.Model) string {
	str := ""
	for _, field := range *s.modelInfo.GetFields() {
		fTag := field.GetFieldTag()
		fType := field.GetFieldType()

		dependType, _ := fType.Depend()
		if dependType != nil {
			continue
		}

		if str == "" {
			str = fmt.Sprintf("`%s`", fTag.Name())
		} else {
			str = fmt.Sprintf("%s,`%s`", str, fTag.Name())
		}
	}

	return str
}
