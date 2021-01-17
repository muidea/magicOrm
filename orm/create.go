package orm

import (
	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *Orm) createSchema(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	tableName := builder.GetTableName()

	existFlag, existErr := s.executor.CheckTableExist(tableName)
	if existErr != nil {
		err = existErr
		return
	}

	if !existFlag {
		// no exist
		sql, err := builder.BuildCreateSchema()
		if err != nil {
			log.Errorf("build create schema failed, err:%s", err.Error())
			return err
		}

		_, err = s.executor.Execute(sql)
	}

	return
}

func (s *Orm) createRelationSchema(modelInfo model.Model, fieldName string, relationInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	tableName := builder.GetRelationTableName(fieldName, relationInfo)

	existFlag, existErr := s.executor.CheckTableExist(tableName)
	if existErr != nil {
		err = existErr
		return
	}
	if !existFlag {
		// no exist
		sql, err := builder.BuildCreateRelationSchema(fieldName, relationInfo)
		if err != nil {
			return err
		}

		_, err = s.executor.Execute(sql)
	}

	return
}

func (s *Orm) batchCreateSchema(modelInfo model.Model) (err error) {
	err = s.createSchema(modelInfo)
	if err != nil {
		return
	}

	for _, field := range modelInfo.GetFields() {
		fType := field.GetType()
		if fType.IsBasic() {
			continue
		}

		relationInfo, relationErr := s.modelProvider.GetTypeModel(fType)
		if relationErr != nil {
			err = relationErr
			return
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

// Create create
func (s *Orm) Create(entity interface{}) (err error) {
	entityModel, entityErr := s.modelProvider.GetEntityModel(entity)
	if entityErr != nil {
		err = entityErr
		log.Errorf("GetEntityModel failed, err:%s", err.Error())
		return
	}

	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}

	err = s.batchCreateSchema(entityModel)
	if err == nil {
		cErr := s.executor.CommitTransaction()
		if cErr != nil {
			err = cErr
		}
	} else {
		rErr := s.executor.RollbackTransaction()
		if rErr != nil {
			err = rErr
		}
	}

	return
}
