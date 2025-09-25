package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/model"
)

// BuildInsert  Build Insert
func (s *Builder) BuildInsert(vModel model.Model) (ret database.Result, err *cd.Error) {
	resultStackPtr := &ResultStack{}
	fieldNames := ""
	fieldValues := ""
	for _, field := range vModel.GetFields() {
		if !model.IsBasicField(field) || !model.IsValidField(field) {
			continue
		}

		fSpec := field.GetSpec()
		if fSpec.GetValueDeclare() == model.AutoIncrement {
			continue
		}

		fValue := field.GetValue()
		encodeVal, encodeErr := s.buildCodec.PackedBasicFieldValue(field, fValue)
		if encodeErr != nil {
			err = encodeErr
			log.Errorf("BuildInsert %s failed, encodeFieldValue error:%s", field.GetName(), err.Error())
			return
		}

		if fieldNames == "" {
			fieldNames = fmt.Sprintf("`%s`", field.GetName())
		} else {
			fieldNames = fmt.Sprintf("%s,`%s`", fieldNames, field.GetName())
		}

		resultStackPtr.PushArgs(encodeVal)
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
func (s *Builder) BuildInsertRelation(vModel model.Model, vField model.Field, rModel model.Model) (ret database.Result, err *cd.Error) {
	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildInsertRelation %s failed, s.buildCodec.ConstructRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	leftVal := vModel.GetPrimaryField().GetValue().Get()
	rightVal := rModel.GetPrimaryField().GetValue().Get()
	resultStackPtr := &ResultStack{}
	insertRelationSQL := fmt.Sprintf("INSERT INTO `%s` (`left`, `right`) VALUES (?,?)", relationTableName)
	resultStackPtr.PushArgs(leftVal, rightVal)
	resultStackPtr.SetSQL(insertRelationSQL)
	//log.Print(ret)
	if traceSQL() {
		log.Infof("[SQL] insert relation: %s", insertRelationSQL)
	}

	ret = resultStackPtr
	return
}
