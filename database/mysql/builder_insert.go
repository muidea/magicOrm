package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BuildInsert  Build Insert
func (s *Builder) BuildInsert(vModel model.Model) (ret *ResultStack, err *cd.Result) {
	resultStackPtr := &ResultStack{}
	fieldNames := ""
	fieldValues := ""
	for _, field := range vModel.GetFields() {
		fType := field.GetType()
		fSpec := field.GetSpec()
		if !fType.IsBasic() || fSpec.GetValueDeclare() == model.AutoIncrement {
			continue
		}

		fStr, fErr := s.buildCodec.BuildFieldValue(field)
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

		resultStackPtr.PushArgs(fStr.Value())
		if fieldValues == "" {
			fieldValues = fmt.Sprintf("%v", "?")
		} else {
			fieldValues = fmt.Sprintf("%s,%v", fieldValues, "?")
		}
	}

	insertSQL := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", s.buildCodec.ConstructModelTableName(vModel), fieldNames, fieldValues)
	if traceSQL() {
		log.Infof("[SQL] insert: %s", insertSQL)
	}

	resultStackPtr.SetSQL(insertSQL)
	ret = resultStackPtr
	return
}

// BuildInsertRelation Build Insert Relation
func (s *Builder) BuildInsertRelation(vModel model.Model, vField model.Field, rModel model.Model) (ret *ResultStack, err *cd.Result) {
	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildInsertRelation %s failed, s.buildCodec.ConstructRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	leftVal, leftErr := s.buildCodec.BuildModelValue(vModel)
	if leftErr != nil {
		err = leftErr
		log.Errorf("BuildInsertRelation failed, s.buildCodec.BuildModelValue error:%s", err.Error())
		return
	}

	rightVal, rightErr := s.buildCodec.BuildModelValue(rModel)
	if rightErr != nil {
		err = rightErr
		log.Errorf("BuildInsertRelation failed, s.BuildModelValue error:%s", err.Error())
		return
	}

	resultStackPtr := &ResultStack{}
	insertRelationSQL := fmt.Sprintf("INSERT INTO `%s` (`left`, `right`) VALUES (?,?)", relationTableName)
	resultStackPtr.PushArgs(leftVal.Value(), rightVal.Value())
	resultStackPtr.SetSQL(insertRelationSQL)
	//log.Print(ret)
	if traceSQL() {
		log.Infof("[SQL] insert relation: %s", insertRelationSQL)
	}

	ret = resultStackPtr
	return
}
