package orm

import (
	"fmt"

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
	if fType.IsBasic() {
		return
	}

	relationInfo, relationErr := s.modelProvider.GetTypeModel(fType)
	if relationErr != nil {
		err = relationErr
		return
	}

	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	rightSQL, relationSQL, relationErr := builder.BuildDeleteRelation(fieldInfo.GetName(), relationInfo)
	if relationErr != nil {
		err = relationErr
		return
	}

	elemType := fType.Elem()
	if !elemType.IsPtrType() {
		fieldVal, fieldErr := s.queryRelation(modelInfo, fieldInfo)
		if fieldErr == nil && !fieldVal.IsNil() {
			if util.IsStructType(fType.GetValue()) {
				relationModel, relationErr := s.modelProvider.GetValueModel(fieldVal, fType)
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
					relationModel, relationErr := s.modelProvider.GetValueModel(elemVals[idx], fType.Elem())
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
