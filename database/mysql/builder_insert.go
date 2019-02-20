package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicOrm/model"
)

// BuildInsert  BuildInsert
func (s *Builder) BuildInsert() (ret string, err error) {
	sql := ""
	fieldValues, fieldErr := s.getFieldInsertValues(s.modelInfo)
	if fieldErr != nil {
		err = fieldErr
		return
	}

	fieldNames, fieldErr := s.getFieldInsertNames(s.modelInfo)
	if fieldErr != nil {
		err = fieldErr
		return
	}

	sql = fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", s.getTableName(s.modelInfo), fieldNames, fieldValues)
	log.Print(sql)
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

	ret = fmt.Sprintf("INSERT INTO `%s` (`left`, `right`) VALUES (%s,%s);", s.GetRelationTableName(fieldName, relationInfo), leftVal, rightVal)
	log.Print(ret)

	return
}

func (s *Builder) getFieldInsertNames(info model.Model) (ret string, err error) {
	str := ""
	for _, field := range s.modelInfo.GetFields() {
		fTag := field.GetTag()
		if fTag.IsAutoIncrement() {
			continue
		}

		fType := field.GetType()
		fValue := field.GetValue()
		if fValue == nil {
			continue
		}

		if fType.IsPtrType() && fValue.IsNil() {
			continue
		}

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

func (s *Builder) getFieldInsertValues(info model.Model) (ret string, err error) {
	str := ""
	for _, field := range info.GetFields() {
		fTag := field.GetTag()
		if fTag.IsAutoIncrement() {
			continue
		}

		fType := field.GetType()
		fValue := field.GetValue()
		if fValue == nil {
			continue
		}

		if fType.IsPtrType() && fValue.IsNil() {
			continue
		}

		dependModel, dependErr := s.modelProvider.GetTypeModel(fType.GetType())
		if dependErr != nil {
			err = dependErr
			return
		}
		if dependModel != nil {
			continue
		}

		fStr, ferr := s.modelProvider.GetValueStr(fType, fValue)
		if ferr == nil {
			if str == "" {
				str = fmt.Sprintf("%s", fStr)
			} else {
				str = fmt.Sprintf("%s,%s", str, fStr)
			}
		} else {
			err = ferr
			break
		}
	}

	ret = str

	return
}
