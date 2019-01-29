package orm

import (
	"log"
	"reflect"

	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/local"
	"muidea.com/magicOrm/model"
)

func (s *orm) createSchema(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo)
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
	builder := builder.NewBuilder(modelInfo)
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
	for _, val := range fields {
		fType := val.GetType()
		fDepend := fType.Depend()
		if fDepend == nil {
			continue
		}

		infoVal, infoErr := local.GetTypeModel(fDepend, s.modelInfoCache)
		if infoErr != nil {
			err = infoErr
			return
		}

		if fDepend.Kind() != reflect.Ptr {
			err = s.createSchema(infoVal)
			if err != nil {
				return
			}
		}

		err = s.createRelationSchema(modelInfo, val.GetName(), infoVal)
		if err != nil {
			return
		}
	}

	return
}

func (s *orm) Create(obj interface{}) (err error) {
	modelInfo, structErr := local.GetObjectModel(obj, s.modelInfoCache)
	if structErr != nil {
		err = structErr
		log.Printf("GetObjectModel failed, err:%s", err.Error())
		return
	}

	err = s.batchCreateSchema(modelInfo)
	if err != nil {
		return
	}
	return
}
