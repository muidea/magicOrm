package orm

import (
	"log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
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

	for _, field := range modelInfo.GetFields() {
		fType := field.GetType()
		depend := fType.Depend()
		if depend == nil {
			continue
		}

		if util.IsBasicType(depend.GetValue()) {
			continue
		}

		relationInfo, relationErr := s.modelProvider.GetTypeModel(fType)
		if relationErr != nil {
			err = relationErr
			return
		}
		if relationInfo == nil {
			continue
		}

		if !fType.IsPtrType() {
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

func (s *orm) Create(entity interface{}) (err error) {
	modelInfo, modelErr := s.modelProvider.GetEntityModel(entity)
	if modelErr != nil {
		err = modelErr
		log.Printf("GetEntityModel failed, err:%s", err.Error())
		return
	}

	err = s.batchCreateSchema(modelInfo)
	return
}
