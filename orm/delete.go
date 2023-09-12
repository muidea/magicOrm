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

	_, _, err = s.executor.Execute(sqlStr)
	return
}

func (s *impl) deleteRelationStructInner(rVal model.Value, rType model.Type, deepLevel int) (err error) {
	relationModel, relationErr := s.modelProvider.GetValueModel(rVal, rType)
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

func (s *impl) deleteRelationSliceInner(rVal model.Value, rType model.Type, deepLevel int) (err error) {
	elemVal, elemErr := s.modelProvider.ElemDependValue(rVal)
	if elemErr != nil {
		err = elemErr
		return
	}

	elemType := rType.Elem()
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

func (s *impl) deleteRelation(vModel model.Model, rField model.Field, deepLevel int) (err error) {
	rType := rField.GetType()
	if rType.IsBasic() {
		return
	}

	relationModel, relationErr := s.modelProvider.GetTypeModel(rType)
	if relationErr != nil {
		err = relationErr
		log.Errorf("get relation field model failed, err:%s", err.Error())
		return
	}

	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	rightSQL, relationSQL, buildErr := builder.BuildDeleteRelation(rField, relationModel)
	if buildErr != nil {
		err = buildErr
		return
	}

	elemType := rType.Elem()
	if !elemType.IsPtrType() {
		fieldVal, fieldErr := s.queryRelation(vModel, rField, deepLevel)
		if fieldErr == nil && !fieldVal.IsNil() {
			if model.IsStructType(rType.GetValue()) {
				err = s.deleteRelationStructInner(fieldVal, rType, deepLevel)
				if err != nil {
					return
				}
			} else if model.IsSliceType(rType.GetValue()) {
				err = s.deleteRelationSliceInner(fieldVal, rType, deepLevel)
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
