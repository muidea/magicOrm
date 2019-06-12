package orm

import (
	"log"
	"reflect"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *Orm) updateSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sql, err := builder.BuildUpdate()
	if err != nil {
		return err
	}

	s.executor.Update(sql)

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
	entityVal := reflect.ValueOf(entity)
	modelInfo, modelErr := s.modelProvider.GetValueModel(entityVal)
	if modelErr != nil {
		err = modelErr
		log.Printf("GetValueModel failed, err:%s", err.Error())
		return
	}

	s.executor.BeginTransaction()
	for {
		err = s.updateSingle(modelInfo)
		if err != nil {
			break
		}

		for _, field := range modelInfo.GetFields() {
			err = s.updateRelation(modelInfo, field)
			if err != nil {
				break
			}
		}

		break
	}

	if err == nil {
		s.executor.CommitTransaction()
	} else {
		s.executor.RollbackTransaction()
	}

	return
}
