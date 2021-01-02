package orm

import (
	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
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

	typeVal := pk.GetType().Interface(id)
	err = pk.UpdateValue(typeVal)
	if err != nil {
		log.Errorf("UpdateValue failed, err:%s", err.Error())
		return err
	}

	return
}

func (s *Orm) insertRelation(modelInfo model.Model, fieldInfo model.Field) (err error) {
	fValue := fieldInfo.GetValue()
	if fValue.IsNil() {
		return
	}

	fType := fieldInfo.GetType()
	fSliceValue, fSliceErr := s.modelProvider.ElemDependValue(fValue)
	if fSliceErr != nil {
		err = fSliceErr
		log.Errorf("ElemDependValue failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
		return
	}

	for _, fVal := range fSliceValue {
		relationInfo, relationErr := s.modelProvider.GetValueModel(fVal, fType)
		if relationErr != nil {
			log.Errorf("GetValueModel failed, err:%s", relationErr.Error())
			err = relationErr
			return
		}

		elemType := fType.Elem()
		if !elemType.IsPtrType() {
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
		err = s.insertSingle(entityModel)
		if err != nil {
			log.Errorf("insertSingle failed, name:%s, err:%s", entityModel.GetName(), err.Error())
			break
		}

		for _, field := range entityModel.GetFields() {
			err = s.insertRelation(entityModel, field)
			if err != nil {
				log.Errorf("insertRelation failed, name:%s, field:%s, err:%s", entityModel.GetName(), field.GetName(), err.Error())
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
