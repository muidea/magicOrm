package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// BuildInsert  Build Insert
func (s *Builder) BuildInsert() (ret string, err error) {
	fieldNames := ""
	fieldValues := ""
	for _, field := range s.GetFields() {
		fType := field.GetType()
		fSpec := field.GetSpec()
		fValue := field.GetValue()
		if !fType.IsBasic() || fValue.IsNil() || fSpec.GetValueDeclare() == model.AutoIncrement {
			continue
		}

		valStr, valErr := s.EncodeValue(fValue, fType)
		if valErr != nil {
			err = valErr
			return
		}

		if fieldNames == "" {
			fieldNames = fmt.Sprintf("`%s`", field.GetName())
		} else {
			fieldNames = fmt.Sprintf("%s,`%s`", fieldNames, field.GetName())
		}

		if fieldValues == "" {
			fieldValues = fmt.Sprintf("%v", valStr)
		} else {
			fieldValues = fmt.Sprintf("%s,%v", fieldValues, valStr)
		}
	}

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", s.GetTableName(), fieldNames, fieldValues)
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
