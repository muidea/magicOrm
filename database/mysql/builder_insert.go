package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/model"
)

// BuildInsert  Build Insert
func (s *Builder) BuildInsert() (ret codec.BuildResult, err *cd.Result) {
	fieldNames := ""
	fieldValues := ""
	for _, field := range s.hostModel.GetFields() {
		fType := field.GetType()
		fSpec := field.GetSpec()
		fValue := field.GetValue()
		if !fType.IsBasic() || fSpec.GetValueDeclare() == model.AutoIncrement {
			continue
		}

		fStr, fErr := s.buildContext.BuildFieldValue(fType, fValue)
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

	insertSQL := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", s.buildContext.BuildHostModelTableName(), fieldNames, fieldValues)
	//log.Print(insertSQL)
	if traceSQL() {
		log.Infof("[SQL] insert: %s", insertSQL)
	}

	ret = NewBuildResult(insertSQL, nil)
	return
}

// BuildInsertRelation Build Insert Relation
func (s *Builder) BuildInsertRelation(vField model.Field, rModel model.Model) (ret codec.BuildResult, err *cd.Result) {
	leftVal, rightVal, valErr := s.buildContext.BuildRelationValue(rModel)
	if valErr != nil {
		err = valErr
		log.Errorf("BuildInsertRelation failed, s.BuildRelationValue error:%s", err.Error())
		return
	}
	relationTableName, relationErr := s.buildContext.BuildRelationTableName(vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildInsertRelation %s failed, s.buildContext.BuildRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	insertRelationSQL := fmt.Sprintf("INSERT INTO `%s` (`left`, `right`) VALUES (%v,%v)", relationTableName, leftVal, rightVal)
	//log.Print(ret)
	if traceSQL() {
		log.Infof("[SQL] insert relation: %s", insertRelationSQL)
	}

	ret = NewBuildResult(insertRelationSQL, nil)
	return
}
