package postgres

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
)

// BuildInsert  Build Insert
func (s *Builder) BuildInsert(vModel models.Model) (ret database.Result, err *cd.Error) {
	resultStackPtr := &ResultStack{}
	fieldNames := ""
	fieldValues := ""
	pkName := ""
	for _, field := range vModel.GetFields() {
		fSpec := field.GetSpec()
		// 检查 req 约束
		constraints := fSpec.GetConstraints()
		if constraints != nil && (constraints.Has(models.KeyRequired) || constraints.Has(models.KeyImmutable)) {
			if !models.IsAssignedField(field) {
				err = cd.NewError(cd.IllegalParam, fmt.Sprintf("field '%s' is required, cannot be zero value", field.GetName()))
				log.Errorf("BuildInsert failed, required field '%s' is zero value", field.GetName())
				return
			}
		}

		if models.IsPrimaryField(field) {
			pkName = field.GetName()
		}
		if !models.IsBasicField(field) || !models.IsValidField(field) {
			continue
		}

		if fSpec.GetValueDeclare() == models.AutoIncrement {
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
			fieldNames = fmt.Sprintf("\"%s\"", field.GetName())
		} else {
			fieldNames = fmt.Sprintf("%s,\"%s\"", fieldNames, field.GetName())
		}

		resultStackPtr.PushArgs(encodeVal)
		if fieldValues == "" {
			fieldValues = fmt.Sprintf("$%d", len(resultStackPtr.argsVal))
		} else {
			fieldValues = fmt.Sprintf("%s,$%d", fieldValues, len(resultStackPtr.argsVal))
		}
	}

	insertSQL := fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES (%s) RETURNING %s", s.buildCodec.ConstructModelTableName(vModel), fieldNames, fieldValues, pkName)
	if traceSQL() {
		log.Infof("[SQL] insert: %s", insertSQL)
	}

	resultStackPtr.SetSQL(insertSQL)
	ret = resultStackPtr
	return
}

// BuildInsertRelation Build Insert Relation
func (s *Builder) BuildInsertRelation(vModel models.Model, vField models.Field, rModel models.Model) (ret database.Result, err *cd.Error) {
	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildInsertRelation %s failed, s.buildCodec.ConstructRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	leftVal := vModel.GetPrimaryField().GetValue().Get()
	rightVal := rModel.GetPrimaryField().GetValue().Get()
	resultStackPtr := &ResultStack{}
	insertRelationSQL := fmt.Sprintf("INSERT INTO \"%s\" (\"left\", \"right\") VALUES ($1,$2) RETURNING id", relationTableName)
	resultStackPtr.PushArgs(leftVal, rightVal)
	resultStackPtr.SetSQL(insertRelationSQL)
	//log.Print(ret)
	if traceSQL() {
		log.Infof("[SQL] insert relation: %s", insertRelationSQL)
	}

	ret = resultStackPtr
	return
}
