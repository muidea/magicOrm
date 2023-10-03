package orm

import (
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) deleteSingle(vModel model.Model) (err error) {
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	sqlStr, sqlErr := builder.BuildDelete()
	if sqlErr != nil {
		err = sqlErr
		log.Errorf("deleteSingle failed, builder.BuildDelete error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(sqlStr)
	if err != nil {
		log.Errorf("deleteSingle failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *impl) deleteRelationStructInner(rVal model.Value, rType model.Type, deepLevel int) (err error) {
	relationModel, relationErr := s.modelProvider.GetValueModel(rVal, rType)
	if relationErr != nil {
		err = relationErr
		log.Errorf("deleteRelationStructInner failed, s.modelProvider.GetValueModel error:%s", err.Error())
		return
	}

	err = s.deleteSingle(relationModel)
	if err != nil {
		log.Errorf("deleteRelationStructInner failed, s.deleteSingle error:%s", err.Error())
		return
	}

	for _, field := range relationModel.GetFields() {
		if field.IsBasic() {
			continue
		}

		err = s.deleteRelation(relationModel, field, deepLevel+1)
		if err != nil {
			log.Errorf("deleteRelationStructInner failed, s.deleteRelation error:%s", err.Error())
			return
		}
	}

	return
}

func (s *impl) deleteRelationSliceInner(rVal model.Value, rType model.Type, deepLevel int) (err error) {
	elemVal, elemErr := s.modelProvider.ElemDependValue(rVal)
	if elemErr != nil {
		err = elemErr
		log.Errorf("deleteRelationSliceInner failed, s.modelProvider.ElemDependValue error:%s", err.Error())
		return
	}

	elemType := rType.Elem()
	for idx := 0; idx < len(elemVal); idx++ {
		relationModel, relationErr := s.modelProvider.GetValueModel(elemVal[idx], elemType)
		if relationErr != nil {
			err = relationErr
			log.Errorf("deleteRelationSliceInner failed, s.modelProvider.GetValueModel error:%s", err.Error())
			return
		}

		err = s.deleteSingle(relationModel)
		if err != nil {
			log.Errorf("deleteRelationSliceInner failed, s.deleteSingle error:%s", err.Error())
			return
		}

		for _, field := range relationModel.GetFields() {
			if field.IsBasic() {
				continue
			}

			err = s.deleteRelation(relationModel, field, deepLevel+1)
			if err != nil {
				log.Errorf("deleteRelationSliceInner failed, s.deleteRelation error:%s", err.Error())
				return
			}
		}
	}

	return
}

func (s *impl) deleteRelation(vModel model.Model, rField model.Field, deepLevel int) (err error) {
	rType := rField.GetType()
	rValue := rField.GetValue()
	relationModel, relationErr := s.modelProvider.GetValueModel(rValue, rType)
	if relationErr != nil {
		err = relationErr
		log.Errorf("deleteRelation failed, s.modelProvider.GetTypeModel error:%s", err.Error())
		return
	}

	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	rightSQL, relationSQL, buildErr := builder.BuildDeleteRelation(rField, relationModel)
	if buildErr != nil {
		err = buildErr
		log.Errorf("deleteRelation failed, builder.BuildDeleteRelation error:%s", err.Error())
		return
	}

	elemType := rType.Elem()
	if !elemType.IsPtrType() {
		fieldVal, fieldErr := s.queryRelation(vModel, rField, maxDeepLevel-1)
		if fieldErr != nil {
			err = fieldErr
			log.Errorf("deleteRelation failed, s.queryRelation error:%s", err.Error())
			return
		}

		if model.IsStructType(rType.GetValue()) {
			err = s.deleteRelationStructInner(fieldVal, rType, deepLevel)
			if err != nil {
				log.Errorf("deleteRelation failed, s.deleteRelationStructInner error:%s", err.Error())
				return
			}
		} else if model.IsSliceType(rType.GetValue()) {
			err = s.deleteRelationSliceInner(fieldVal, rType, deepLevel)
			if err != nil {
				log.Errorf("deleteRelation failed, s.deleteRelationSliceInner error:%s", err.Error())
				return
			}
		}

		_, _, err = s.executor.Execute(rightSQL)
		if err != nil {
			log.Errorf("deleteRelation failed, s.executor.Execute error:%s", err.Error())
			return
		}
	}

	_, _, err = s.executor.Execute(relationSQL)
	if err != nil {
		log.Errorf("deleteRelation failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

// Delete delete
func (s *impl) Delete(vModel model.Model) (ret model.Model, err error) {
	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer func() {
		err = s.finalTransaction(err)
	}()

	err = s.deleteSingle(vModel)
	if err != nil {
		log.Errorf("Delete failed, s.deleteSingle error:%s", err.Error())
		return
	}

	for _, field := range vModel.GetFields() {
		if field.IsBasic() {
			continue
		}

		err = s.deleteRelation(vModel, field, 0)
		if err != nil {
			log.Errorf("Delete failed, s.deleteRelation error:%s", err.Error())
			return
		}
	}

	ret = vModel
	return
}
