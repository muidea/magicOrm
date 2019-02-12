package orm

import (
	"log"

	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/model"
)

func (s *orm) createSchema(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	tableName := builder.GetTableName()

	if !s.executor.CheckTableExist(tableName) {
		// no exist
		sql, err := builder.BuildCreateSchema()
		if err != nil {
			log.Printf("build create schema failed, err:%s", err.Error())
			return err
		}

		s.executor.Execute(sql)
	}

	return
}

func (s *orm) createRelationSchema(modelInfo model.Model, fieldName string, relationInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	tableName := builder.GetRelationTableName(fieldName, relationInfo)

	if !s.executor.CheckTableExist(tableName) {
		// no exist
		sql, err := builder.BuildCreateRelationSchema(fieldName, relationInfo)
		if err != nil {
			return err
		}

		s.executor.Execute(sql)
	}

	return
}

func (s *orm) batchCreateSchema(modelInfo model.Model) (err error) {
	err = s.createSchema(modelInfo)
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

		relationInfo, relationErr := s.modelProvider.GetTypeModel(fDepend.GetType())
		if relationErr != nil {
			err = relationErr
			return
		}

		if !fDepend.IsPtrType() {
			err = s.createSchema(relationInfo)
			if err != nil {
				return
			}
		}

		err = s.createRelationSchema(modelInfo, field.GetName(), relationInfo)
		if err != nil {
			return
		}
	}

	return
}

func (s *orm) Create(obj interface{}) (err error) {
	modelInfo, modelErr := s.modelProvider.GetObjectModel(obj)
	if modelErr != nil {
		err = modelErr
		log.Printf("GetObjectModel failed, err:%s", err.Error())
		return
	}

	err = s.batchCreateSchema(modelInfo)
	return
}
