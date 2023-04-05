package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// BuildInsert  Build Insert
func (s *Builder) BuildInsert() (ret string, err error) {
	sql := ""
	fieldNames, fieldErr := s.getFieldInsertNames(s.entityModel)
	if fieldErr != nil {
		err = fieldErr
		return
	}

	fieldValues, fieldErr := s.getFieldInsertValues(s.entityModel)
	if fieldErr != nil {
		err = fieldErr
		return
	}

	sql = fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", s.GetTableName(), fieldNames, fieldValues)
	//log.Print(sql)
	ret = sql

	return
}

// BuildInsertRelation Build Insert Relation
func (s *Builder) BuildInsertRelation(field model.Field, relationInfo model.Model) (ret string, err error) {
	leftVal, rightVal, valErr := s.getRelationValue(relationInfo)
	if valErr != nil {
		err = valErr
		return
	}
	relationSchema := s.GetRelationTableName(field, relationInfo)
	ret = fmt.Sprintf("INSERT INTO `%s` (`left`, `right`) VALUES (%v,%v);", relationSchema, leftVal, rightVal)
	//log.Print(ret)

	return
}

func (s *Builder) getFieldInsertNames(info model.Model) (ret string, err error) {
	str := ""
	for _, field := range info.GetFields() {
		fTag := field.GetTag()
		if fTag.IsAutoIncrement() {
			continue
		}

		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() || fValue.IsNil() {
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

func (s *Builder) getFieldInsertValues(info model.Model) (ret string, err error) {
	str := ""
	for _, field := range info.GetFields() {
		fTag := field.GetTag()
		if fTag.IsAutoIncrement() {
			continue
		}

		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() || fValue.IsNil() {
			continue
		}

		fStr, fErr := s.encodeValue(fValue, fType)
		if fErr != nil {
			err = fErr
			return
		}

		if str == "" {
			str = fmt.Sprintf("%v", fStr)
		} else {
			str = fmt.Sprintf("%s,%v", str, fStr)
		}
	}

	ret = str

	return
}
