package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) deleteSingle(vModel model.Model) (err *cd.Result) {
	builderVal := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	sqlStr, sqlErr := builderVal.BuildDelete()
	if sqlErr != nil {
		err = sqlErr
		log.Errorf("deleteSingle failed, builderVal.BuildDelete error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(sqlStr)
	if err != nil {
		log.Errorf("deleteSingle failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *impl) deleteRelationSingleStructInner(rVal model.Value, rType model.Type, deepLevel int) (err *cd.Result) {
	relationModel, relationErr := s.modelProvider.GetValueModel(rVal, rType)
	if relationErr != nil {
		err = relationErr
		log.Errorf("deleteRelationSingleStructInner failed, s.modelProvider.GetValueModel error:%s", err.Error())
		return
	}

	err = s.deleteSingle(relationModel)
	if err != nil {
		log.Errorf("deleteRelationSingleStructInner failed, s.deleteSingle error:%s", err.Error())
		return
	}

	for _, field := range relationModel.GetFields() {
		if field.IsBasic() {
			continue
		}

		err = s.deleteRelation(relationModel, field, deepLevel+1)
		if err != nil {
			log.Errorf("deleteRelationSingleStructInner failed, s.deleteRelation error:%s", err.Error())
			return
		}
	}

	return
}

func (s *impl) deleteRelationSliceStructInner(rVal model.Value, rType model.Type, deepLevel int) (err *cd.Result) {
	elemVal, elemErr := s.modelProvider.ElemDependValue(rVal)
	if elemErr != nil {
		err = elemErr
		log.Errorf("deleteRelationSliceStructInner failed, s.modelProvider.ElemDependValue error:%s", err.Error())
		return
	}

	elemType := rType.Elem()
	for idx := 0; idx < len(elemVal); idx++ {
		err = s.deleteRelationSingleStructInner(elemVal[idx], elemType, deepLevel)
		if err != nil {
			log.Errorf("deleteRelationSliceStructInner failed, s.deleteRelationSingleStructInner error:%s", err.Error())
			return
		}
	}

	return
}

func (s *impl) deleteRelation(vModel model.Model, rField model.Field, deepLevel int) (err *cd.Result) {
	rType := rField.GetType()
	rModel, rErr := s.modelProvider.GetTypeModel(rType)
	if rErr != nil {
		err = rErr
		log.Errorf("deleteRelation failed, s.modelProvider.GetTypeModel error:%s", err.Error())
		return
	}

	builderVal := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	rightSQL, relationSQL, buildErr := builderVal.BuildDeleteRelation(rField, rModel)
	if buildErr != nil {
		err = buildErr
		log.Errorf("deleteRelation failed, builderVal.BuildDeleteRelation error:%s", err.Error())
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
			err = s.deleteRelationSingleStructInner(fieldVal, rType, deepLevel)
			if err != nil {
				log.Errorf("deleteRelation failed, s.deleteRelationSingleStructInner error:%s", err.Error())
				return
			}
		} else if model.IsSliceType(rType.GetValue()) {
			err = s.deleteRelationSliceStructInner(fieldVal, rType, deepLevel)
			if err != nil {
				log.Errorf("deleteRelation failed, s.deleteRelationSliceStructInner error:%s", err.Error())
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
func (s *impl) Delete(vModel model.Model) (ret model.Model, err *cd.Result) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}

	defer s.finalTransaction(err)

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
