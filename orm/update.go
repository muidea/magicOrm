package orm

import (
	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *Orm) updateSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sql, err := builder.BuildUpdate()
	if err != nil {
		return err
	}

	_, err = s.executor.Update(sql)

	return err
}

func (s *Orm) updateRelation(modelInfo model.Model, fieldInfo model.Field) (err error) {
	fType := fieldInfo.GetType()
	fDependModel, fDependErr := s.modelProvider.GetTypeModel(fType)
	if fDependErr != nil {
		err = fDependErr
		return
	}
	if fDependModel == nil {
		return
	}

	err = s.deleteRelation(modelInfo, fieldInfo)
	if err != nil {
		return
	}

	err = s.insertRelation(modelInfo, fieldInfo)
	if err != nil {
		return
	}

	return
}

// Update update
func (s *Orm) Update(entity interface{}) (err error) {
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

	for {
		err = s.updateSingle(entityModel)
		if err != nil {
			break
		}

		for _, field := range entityModel.GetFields() {
			err = s.updateRelation(entityModel, field)
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
