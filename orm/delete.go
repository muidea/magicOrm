package orm

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func (s *Orm) deleteSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sql, err := builder.BuildDelete()
	if err != nil {
		return err
	}
	num, numErr := s.executor.Delete(sql)
	if numErr != nil {
		err = numErr
		return
	}

	if num != 1 {
		log.Errorf("unexception delete, rowNum:%d", num)
		err = fmt.Errorf("delete %s failed", modelInfo.GetName())
	}

	return
}

func (s *Orm) deleteRelation(modelInfo model.Model, fieldInfo model.Field) (err error) {
	fType := fieldInfo.GetType()
	relationInfo, relationErr := s.modelProvider.GetTypeModel(fType)
	if relationErr != nil {
		err = relationErr
		return
	}
	if relationInfo == nil {
		return
	}

	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	rightSQL, relationSQL, relationErr := builder.BuildDeleteRelation(fieldInfo.GetName(), relationInfo)
	if relationErr != nil {
		err = relationErr
		return
	}

	dependType := fType.Depend()
	if !dependType.IsPtrType() {
		fieldVal, fieldErr := s.queryRelation(modelInfo, fieldInfo)
		fieldVal = reflect.Indirect(fieldVal)
		if fieldErr == nil && !util.IsNil(fieldVal) {
			if util.IsStructType(fType.GetValue()) {
				relationModel, relationErr := s.modelProvider.GetValueModel(fieldVal)
				if relationErr != nil {
					err = relationErr
					return
				}

				err = s.deleteSingle(relationModel)
				if err != nil {
					return
				}

				for _, field := range relationModel.GetFields() {
					err = s.deleteRelation(relationModel, field)
					if err != nil {
						break
					}
				}
			} else if util.IsSliceType(fType.GetValue()) {
				elemVals, elemErr := s.modelProvider.ElemDependValue(fieldVal)
				if elemErr != nil {
					err = elemErr
					return
				}
				for idx := 0; idx < len(elemVals); idx++ {
					relationModel, relationErr := s.modelProvider.GetValueModel(elemVals[idx])
					if relationErr != nil {
						err = relationErr
						return
					}

					err = s.deleteSingle(relationModel)
					if err != nil {
						return
					}

					for _, field := range relationModel.GetFields() {
						err = s.deleteRelation(relationModel, field)
						if err != nil {
							break
						}
					}
				}
			}
		}

		_, err = s.executor.Delete(rightSQL)
		if err != nil {
			return
		}
	}

	_, err = s.executor.Delete(relationSQL)

	return
}

// Delete delete
func (s *Orm) Delete(entity interface{}) (err error) {
	entityVal := reflect.ValueOf(entity)
	entityVal = reflect.Indirect(entityVal)
	entityModel, entityErr := s.modelProvider.GetValueModel(entityVal)
	if entityErr != nil {
		err = entityErr
		log.Errorf("GetValueModel failed, err:%s", err.Error())
		return
	}

	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}

	for {
		err = s.deleteSingle(entityModel)
		if err != nil {
			break
		}

		for _, field := range entityModel.GetFields() {
			err = s.deleteRelation(entityModel, field)
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
