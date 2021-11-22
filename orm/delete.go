package orm

import (
	"fmt"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func (s *impl) deleteSingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sqlStr, sqlErr := builder.BuildDelete()
	if sqlErr != nil {
		err = sqlErr
		return
	}

	numVal, numErr := s.executor.Delete(sqlStr)
	if numErr != nil {
		err = numErr
		return
	}

	if numVal != 1 {
		err = fmt.Errorf("delete %s failed", modelInfo.GetName())
	}

	return
}

func (s *impl) deleteRelation(modelInfo model.Model, fieldInfo model.Field, deepLevel int) (err error) {
	fType := fieldInfo.GetType()
	if fType.IsBasic() {
		return
	}

	// disable check field value
	//if !s.modelProvider.IsAssigned(fieldInfo.GetValue(), fieldInfo.GetType()) {
	//	return
	//}

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
		fieldVal, fieldErr := s.queryRelation(modelInfo, fieldInfo, deepLevel)
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
					err = s.deleteRelation(relationModel, field, deepLevel+1)
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
						err = s.deleteRelation(relationModel, field, deepLevel+1)
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
func (s *impl) Delete(entityModel model.Model) (ret model.Model, err error) {
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
			err = s.deleteRelation(entityModel, field, 0)
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

	if err == nil {
		ret = entityModel
	}

	return
}
