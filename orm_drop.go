package orm

import (
	"log"

	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/model"
)

func (s *orm) dropSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo)
	tableName := builder.GetTableName()
	if s.executor.CheckTableExist(tableName) {
		sql, err := builder.BuildDropSchema()
		if err != nil {
			return err
		}

		s.executor.Execute(sql)
	}

	return
}

func (s *orm) dropRelation(modelInfo model.Model, fieldName string, relationInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo)
	tableName := builder.GetRelationTableName(fieldName, relationInfo)
	if s.executor.CheckTableExist(tableName) {
		sql, err := builder.BuildDropRelationSchema(fieldName, relationInfo)
		if err != nil {
			return err
		}

		s.executor.Execute(sql)
	}

	return
}

func (s *orm) Drop(obj interface{}) (err error) {
	modelInfo, structErr := model.GetObjectStructInfo(obj, s.modelInfoCache)
	if structErr != nil {
		err = structErr
		log.Printf("GetObjectStructInfo failed, err:%s", err.Error())
		return
	}

	err = s.dropSingle(modelInfo)
	if err != nil {
		return
	}

	fields := modelInfo.GetDependField()
	for _, val := range fields {
		fType := val.GetFieldType()
		fDepend, fDependPtr := fType.Depend()
		if fDepend == nil {
			continue
		}

		infoVal, infoErr := model.GetStructInfo(fDepend, s.modelInfoCache)
		if infoErr != nil {
			err = infoErr
			return
		}

		if !fDependPtr {
			err = s.dropSingle(infoVal)
			if err != nil {
				return
			}
		}

		err = s.dropRelation(modelInfo, val.GetFieldName(), infoVal)
		if err != nil {
			return
		}
	}

	return
}
