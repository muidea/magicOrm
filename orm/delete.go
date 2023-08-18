package orm

import (
	log "github.com/cihub/seelog"
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) deleteSingle(vModel model.Model) (err error) {
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	sqlStr, sqlErr := builder.BuildDelete()
	if sqlErr != nil {
		err = sqlErr
		return
	}

	_, _, numErr := s.executor.Execute(sqlStr)
	if numErr != nil {
		err = numErr
		return
	}

	return
}

func (s *impl) deleteRelationStructInner(fieldVal model.Value, relationType model.Type, deepLevel int) (err error) {
	relationModel, relationErr := s.modelProvider.GetValueModel(fieldVal, relationType)
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

	return
}

func (s *impl) deleteRelationSliceInner(fieldVal model.Value, relationType model.Type, deepLevel int) (err error) {
	elemVal, elemErr := s.modelProvider.ElemDependValue(fieldVal)
	if elemErr != nil {
		err = elemErr
		return
	}

	elemType := relationType.Elem()
	for idx := 0; idx < len(elemVal); idx++ {
		relationModel, relationErr := s.modelProvider.GetValueModel(elemVal[idx], elemType)
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

	return
}

func (s *impl) deleteRelation(entityModel model.Model, relationField model.Field, deepLevel int) (err error) {
	relationType := relationField.GetType()
	if relationType.IsBasic() {
		return
	}

	// disable check field value
	//if !s.modelProvider.IsAssigned(relationField.GetValue(), relationField.GetType()) {
	//	return
	//}

	relationModel, relationErr := s.modelProvider.GetTypeModel(relationType)
	if relationErr != nil {
		err = relationErr
		log.Errorf("get relation field model failed, err:%s", err.Error())
		return
	}

	builder := builder.NewBuilder(entityModel, s.modelProvider, s.specialPrefix)
	rightSQL, relationSQL, buildErr := builder.BuildDeleteRelation(relationField, relationModel)
	if buildErr != nil {
		err = buildErr
		return
	}

	elemType := relationType.Elem()
	if !elemType.IsPtrType() {
		fieldVal, fieldErr := s.queryRelation(entityModel, relationField, deepLevel)
		if fieldErr == nil && !fieldVal.IsNil() {
			if model.IsStructType(relationType.GetValue()) {
				err = s.deleteRelationStructInner(fieldVal, relationType, deepLevel)
				if err != nil {
					return
				}
			} else if model.IsSliceType(relationType.GetValue()) {
				err = s.deleteRelationSliceInner(fieldVal, relationType, deepLevel)
				if err != nil {
					return
				}
			}
		}

		_, _, err = s.executor.Execute(rightSQL)
		if err != nil {
			return
		}
	}

	_, _, err = s.executor.Execute(relationSQL)

	return
}

// Delete delete
func (s *impl) Delete(vModel model.Model) (ret model.Model, err error) {
	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer s.finalTransaction(err)

	for {
		err = s.deleteSingle(vModel)
		if err != nil {
			break
		}

		for _, field := range vModel.GetFields() {
			err = s.deleteRelation(vModel, field, 0)
			if err != nil {
				break
			}
		}

		break
	}

	if err != nil {
		return
	}

	ret = vModel
	return
}
