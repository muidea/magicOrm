package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// BuildInsert  BuildInsert
func (s *Builder) BuildInsert() (ret string, err error) {
	sql := ""
	fieldNames, fieldErr := s.getFieldInsertNames(s.modelInfo)
	if fieldErr != nil {
		err = fieldErr
		return
	}

	fieldValues, fieldErr := s.getFieldInsertValues(s.modelInfo)
	if fieldErr != nil {
		err = fieldErr
		return
	}

	sql = fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", s.getHostTableName(s.modelInfo), fieldNames, fieldValues)
	//log.Print(sql)
	ret = sql

	return
}

// BuildInsertRelation BuildInsertRelation
func (s *Builder) BuildInsertRelation(fieldName string, relationInfo model.Model) (ret string, err error) {
	leftVal, rightVal, valErr := s.getRelationValue(relationInfo)
	if valErr != nil {
		err = valErr
		return
	}

	ret = fmt.Sprintf("INSERT INTO `%s` (`left`, `right`) VALUES (%v,%v);", s.GetRelationTableName(fieldName, relationInfo), leftVal, rightVal)
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

		fStr, fErr := s.buildValue(fValue, fType)
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
