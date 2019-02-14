package orm

import (
	"log"
	"reflect"

	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/model"
)

func (s *orm) dropSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
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
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
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
	modelInfo, modelErr := s.modelProvider.GetObjectModel(obj)
	if modelErr != nil {
		err = modelErr
		log.Printf("GetObjectModel failed, err:%s", err.Error())
		return
	}

	err = s.dropSingle(modelInfo)
	if err != nil {
		return
	}

	fields := modelInfo.GetDependField()
	for _, field := range fields {
		fType := field.GetType()
		fDepend := fType.GetDepend()
		if fDepend == nil {
			continue
		}

		relationInfo, relationErr := s.modelProvider.GetTypeModel(fDepend)
		if relationErr != nil {
			err = relationErr
			return
		}

		if fDepend.Kind() != reflect.Ptr {
			err = s.dropSingle(relationInfo)
			if err != nil {
				return
			}
		}

		err = s.dropRelation(modelInfo, field.GetName(), relationInfo)
		if err != nil {
			return
		}
	}

	return
}
