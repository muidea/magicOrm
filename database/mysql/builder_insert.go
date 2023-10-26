package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BuildInsert  Build Insert
func (s *Builder) BuildInsert() (ret string, err *cd.Result) {
	fieldNames := ""
	fieldValues := ""
	for _, field := range s.GetFields() {
		fType := field.GetType()
		fSpec := field.GetSpec()
		fValue := field.GetValue()
		if !fType.IsBasic() || fSpec.GetValueDeclare() == model.AutoIncrement {
			continue
		}

		var valStr string
		var valErr *cd.Result
		if !fValue.IsNil() {
			valStr, valErr = s.EncodeValue(fValue, fType)
		} else {
			valStr, valErr = getTypeDefaultValue(fType)
		}
		if valErr != nil {
			err = valErr
			log.Errorf("BuildInsert failed, s.EncodeValue/getTypeDefaultValue error:%s", valErr.Error())
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

	str := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", s.GetTableName(), fieldNames, fieldValues)
	//log.Print(str)

	ret = str
	return
}

// BuildInsertRelation Build Insert Relation
func (s *Builder) BuildInsertRelation(vField model.Field, rModel model.Model) (ret string, err *cd.Result) {
	leftVal, rightVal, valErr := s.GetRelationValue(rModel)
	if valErr != nil {
		err = valErr
		log.Errorf("BuildInsertRelation failed, s.GetRelationValue error:%s", err.Error())
		return
	}
	relationTableName := s.GetRelationTableName(vField, rModel)
	str := fmt.Sprintf("INSERT INTO `%s` (`left`, `right`) VALUES (%v,%v)", relationTableName, leftVal, rightVal)
	//log.Print(ret)

	ret = str
	return
}
