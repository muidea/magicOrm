package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// BuildInsert  Build Insert
func (s *Builder) BuildInsert() (ret string, err error) {
	sql := ""
	fieldNames, nameErr := s.getFieldInsertNames()
	if nameErr != nil {
		err = nameErr
		return
	}

	fieldValues, valueErr := s.getFieldInsertValues()
	if valueErr != nil {
		err = valueErr
		return
	}

	sql = fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", s.GetTableName(), fieldNames, fieldValues)
	//log.Print(sql)
	ret = sql

	return
}

// BuildInsertRelation Build Insert Relation
func (s *Builder) BuildInsertRelation(vField model.Field, rModel model.Model) (ret string, err error) {
	leftVal, rightVal, valErr := s.GetRelationValue(rModel)
	if valErr != nil {
		err = valErr
		return
	}
	relationSchema := s.GetRelationTableName(vField, rModel)
	ret = fmt.Sprintf("INSERT INTO `%s` (`left`, `right`) VALUES (%v,%v);", relationSchema, leftVal, rightVal)
	//log.Print(ret)

	return
}

func (s *Builder) getFieldInsertNames() (ret string, err error) {
	str := ""
	for _, field := range s.GetFields() {
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

func (s *Builder) getFieldInsertValues() (ret string, err error) {
	str := ""
	for _, field := range s.GetFields() {
		fTag := field.GetTag()
		if fTag.IsAutoIncrement() {
			continue
		}

		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() || fValue.IsNil() {
			continue
		}

		fStr, fErr := s.EncodeValue(fValue, fType)
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
