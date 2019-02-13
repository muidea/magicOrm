package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicOrm/model"
)

// BuildInsert  BuildInsert
func (s *Builder) BuildInsert() (ret string, err error) {
	sql := ""
	val, verr := s.getFieldInsertValues(s.modelInfo)
	if verr == nil {
		sql = fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", s.getTableName(s.modelInfo), s.getFieldInsertNames(s.modelInfo), val)
		log.Print(sql)
		ret = sql
	}
	err = verr

	return
}

// BuildInsertRelation BuildInsertRelation
func (s *Builder) BuildInsertRelation(fieldName string, relationInfo model.Model) (ret string, err error) {
	leftVal, rightVal, errVal := s.getRelationValue(relationInfo)
	if errVal != nil {
		err = errVal
		return
	}

	ret = fmt.Sprintf("INSERT INTO `%s` (`left`, `right`) VALUES (%s,%s);", s.GetRelationTableName(fieldName, relationInfo), leftVal, rightVal)
	log.Print(ret)

	return
}

func (s *Builder) getFieldInsertNames(info model.Model) string {
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

		dependType := fType.GetDepend()
		if dependType != nil {
			continue
		}

		if str == "" {
			str = fmt.Sprintf("`%s`", fTag.GetName())
		} else {
			str = fmt.Sprintf("%s,`%s`", str, fTag.GetName())
		}
	}

	return str
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

		dependType := fType.GetDepend()
		if dependType != nil {
			continue
		}

		fStr, ferr := fValue.GetValueStr()
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