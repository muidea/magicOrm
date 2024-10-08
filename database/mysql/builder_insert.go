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
	for _, field := range s.common.GetHostModelFields() {
		fType := field.GetType()
		fSpec := field.GetSpec()
		fValue := field.GetValue()
		if !fType.IsBasic() || fSpec.GetValueDeclare() == model.AutoIncrement {
			continue
		}

		fStr, fErr := s.common.BuildFieldValue(fType, fValue)
		if fErr != nil {
			err = fErr
			log.Errorf("BuildInsert failed, BuildFieldValue error:%s", fErr.Error())
			return
		}

		if fieldNames == "" {
			fieldNames = fmt.Sprintf("`%s`", field.GetName())
		} else {
			fieldNames = fmt.Sprintf("%s,`%s`", fieldNames, field.GetName())
		}

		if fieldValues == "" {
			fieldValues = fmt.Sprintf("%v", fStr)
		} else {
			fieldValues = fmt.Sprintf("%s,%v", fieldValues, fStr)
		}
	}

	str := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", s.common.BuildHostModelTableName(), fieldNames, fieldValues)
	//log.Print(str)
	if traceSQL() {
		log.Infof("[SQL] insert: %s", str)
	}

	ret = str
	return
}

// BuildInsertRelation Build Insert Relation
func (s *Builder) BuildInsertRelation(vField model.Field, rModel model.Model) (ret string, err *cd.Result) {
	leftVal, rightVal, valErr := s.common.BuildRelationValue(rModel)
	if valErr != nil {
		err = valErr
		log.Errorf("BuildInsertRelation failed, s.BuildRelationValue error:%s", err.Error())
		return
	}
	relationTableName, relationErr := s.common.BuildRelationTableName(vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildInsertRelation %s failed, s.common.BuildRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	str := fmt.Sprintf("INSERT INTO `%s` (`left`, `right`) VALUES (%v,%v)", relationTableName, leftVal, rightVal)
	//log.Print(ret)
	if traceSQL() {
		log.Infof("[SQL] insert relation: %s", str)
	}

	ret = str
	return
}
