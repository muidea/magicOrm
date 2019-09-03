package orm

import (
	"log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *Orm) dropSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	tableName := builder.GetTableName()

	existFlag, existErr := s.executor.CheckTableExist(tableName)
	if existErr != nil {
		err = existErr
		return
	}

	if existFlag {
		sql, err := builder.BuildDropSchema()
		if err != nil {
			return err
		}

		_, err = s.executor.Execute(sql)
	}

	return
}

func (s *Orm) dropRelation(modelInfo model.Model, fieldName string, relationInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	tableName := builder.GetRelationTableName(fieldName, relationInfo)

	existFlag, existErr := s.executor.CheckTableExist(tableName)
	if existErr != nil {
		err = existErr
		return
	}
	if existFlag {
		sql, err := builder.BuildDropRelationSchema(fieldName, relationInfo)
		if err != nil {
			return err
		}

		_, err = s.executor.Execute(sql)
	}

	return
}

// Drop drop
func (s *Orm) Drop(entity interface{}) (err error) {
	modelInfo, modelErr := s.modelProvider.GetEntityModel(entity)
	if modelErr != nil {
		err = modelErr
		log.Printf("GetEntityModel failed, err:%s", err.Error())
		return
	}

	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}

	for {
		err = s.dropSingle(modelInfo)
		if err != nil {
			break
		}

		for _, field := range modelInfo.GetFields() {
			fType := field.GetType()
			relationInfo, relationErr := s.modelProvider.GetTypeModel(fType)
			if relationErr != nil {
				err = relationErr
				break
			}
			if relationInfo == nil {
				continue
			}

			if !relationInfo.IsPtrModel() {
				err = s.dropSingle(relationInfo)
				if err != nil {
					break
				}
			}

			err = s.dropRelation(modelInfo, field.GetName(), relationInfo)
			if err != nil {
				break
			}
		}
		break
	}

	if err == nil {
		err = s.executor.CommitTransaction()
	} else {
		err = s.executor.RollbackTransaction()
	}

	return
}
