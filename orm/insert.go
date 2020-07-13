package orm

import (
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/helper"
)

func (s *Orm) insertSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sql, err := builder.BuildInsert()
	if err != nil {
		log.Errorf("BuildInsert failed, err:%s", err.Error())
		return err
	}

	id, idErr := s.executor.Insert(sql)
	if idErr != nil {
		err = idErr
		return
	}
	pk := modelInfo.GetPrimaryField()
	if pk != nil {
		typeVal := pk.GetType().Interface()
		typeVal, err = helper.AssignValue(reflect.ValueOf(id), typeVal)
		if err != nil {
			log.Errorf("assign pk field failed, err:%s", err.Error())
			return err
		}

		err = pk.UpdateValue(typeVal)
		if err != nil {
			log.Errorf("UpdateValue failed, err:%s", err.Error())
			return err
		}
	}

	return
}

func (s *Orm) insertRelation(modelInfo model.Model, fieldInfo model.Field) (err error) {
	fType := fieldInfo.GetType()
	fDependModel, fDependErr := s.modelProvider.GetTypeModel(fType)
	if fDependErr != nil {
		err = fDependErr
		return
	}
	if fDependModel == nil {
		return
	}

	fDependValue, fDependErr := s.modelProvider.GetModelDependValue(fDependModel, fieldInfo.GetValue())
	if fDependErr != nil {
		err = fDependErr
		log.Errorf("GetModelDependValue failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
		return
	}

	for _, fVal := range fDependValue {
		relationInfo, relationErr := s.modelProvider.GetValueModel(fVal)
		if relationErr != nil {
			log.Errorf("GetValueModel faield, err:%s", relationErr.Error())
			err = relationErr
			return
		}

		if !relationInfo.IsPtrModel() {
			err = s.insertSingle(relationInfo)
			if err != nil {
				return
			}

			for _, subField := range relationInfo.GetFields() {
				err = s.insertRelation(relationInfo, subField)
				if err != nil {
					return
				}
			}
		}

		builder := builder.NewBuilder(modelInfo, s.modelProvider)
		relationSQL, relationErr := builder.BuildInsertRelation(fieldInfo.GetName(), relationInfo)
		if relationErr != nil {
			err = relationErr
			return err
		}

		_, err = s.executor.Insert(relationSQL)
		if err != nil {
			return
		}
	}

	return
}

// Insert insert
func (s *Orm) Insert(entity interface{}) (err error) {
	entityVal := reflect.ValueOf(entity)
	modelInfo, modelErr := s.modelProvider.GetValueModel(entityVal)
	if modelErr != nil {
		err = modelErr
		log.Errorf("GetValueModel failed, err:%s", err.Error())
		return
	}

	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}

	for {
		err = s.insertSingle(modelInfo)
		if err != nil {
			log.Errorf("insertSingle failed, name:%s, err:%s", modelInfo.GetName(), err.Error())
			break
		}

		for _, field := range modelInfo.GetFields() {
			err = s.insertRelation(modelInfo, field)
			if err != nil {
				log.Errorf("insertRelation failed, name:%s, field:%s, err:%s", modelInfo.GetName(), field.GetName(), err.Error())
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
