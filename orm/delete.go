package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

type DeleteRunner struct {
	baseRunner
	QueryRunner
}

func NewDeleteRunner(
	vModel model.Model,
	executor executor.Executor,
	provider provider.Provider,
	modelCodec codec.Codec,
	deepLevel int) *DeleteRunner {
	baseRunner := newBaseRunner(vModel, executor, provider, modelCodec, false, deepLevel)
	return &DeleteRunner{
		baseRunner: baseRunner,
		QueryRunner: QueryRunner{
			baseRunner: baseRunner,
		},
	}
}

func (s *DeleteRunner) deleteHost() (err *cd.Result) {
	deleteResult, deleteErr := s.hBuilder.BuildDelete()
	if deleteErr != nil {
		err = deleteErr
		log.Errorf("deleteHost failed, builderVal.BuildDelete error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(deleteResult.SQL(), deleteResult.Args()...)
	if err != nil {
		log.Errorf("deleteHost failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *DeleteRunner) deleteRelation(vField model.Field, deepLevel int) (err *cd.Result) {
	vType := vField.GetType()
	rModel, rErr := s.modelProvider.GetTypeModel(vType)
	if rErr != nil {
		err = rErr
		log.Errorf("deleteRelation failed, s.modelProvider.GetTypeModel error:%s", err.Error())
		return
	}

	rightResult, relationResult, resultErr := s.hBuilder.BuildDeleteRelation(vField, rModel)
	if resultErr != nil {
		err = resultErr
		log.Errorf("deleteRelation failed, s.hBuilder.BuildDeleteRelation error:%s", err.Error())
		return
	}

	elemType := vType.Elem()
	if !elemType.IsPtrType() {
		fieldVal, fieldErr := s.queryRelation(vField, maxDeepLevel-1)
		if fieldErr != nil {
			err = fieldErr
			log.Errorf("deleteRelation failed, s.queryRelation error:%s", err.Error())
			return
		}

		if model.IsStructType(vType.GetValue()) {
			err = s.deleteRelationSingleStructInner(fieldVal, vType, deepLevel)
			if err != nil {
				log.Errorf("deleteRelation failed, s.deleteRelationSingleStructInner error:%s", err.Error())
				return
			}
		} else if model.IsSliceType(vType.GetValue()) {
			err = s.deleteRelationSliceStructInner(fieldVal, vType, deepLevel)
			if err != nil {
				log.Errorf("deleteRelation failed, s.deleteRelationSliceStructInner error:%s", err.Error())
				return
			}
		}

		_, _, err = s.executor.Execute(rightResult.SQL(), rightResult.Args()...)
		if err != nil {
			log.Errorf("deleteRelation failed, s.executor.Execute error:%s", err.Error())
			return
		}
	}

	_, _, err = s.executor.Execute(relationResult.SQL(), relationResult.Args()...)
	if err != nil {
		log.Errorf("deleteRelation failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *DeleteRunner) deleteRelationSingleStructInner(rVal model.Value, rType model.Type, deepLevel int) (err *cd.Result) {
	relationModel, relationErr := s.modelProvider.GetValueModel(rVal, rType)
	if relationErr != nil {
		err = relationErr
		log.Errorf("deleteRelationSingleStructInner failed, s.modelProvider.GetValueModel error:%s", err.Error())
		return
	}

	rRunner := NewDeleteRunner(relationModel, s.executor, s.modelProvider, s.modelCodec, deepLevel+1)
	err = rRunner.Delete()
	if err != nil {
		log.Errorf("deleteRelationSingleStructInner failed, s.deleteSingle error:%s", err.Error())
		return
	}

	return
}

func (s *DeleteRunner) deleteRelationSliceStructInner(rVal model.Value, rType model.Type, deepLevel int) (err *cd.Result) {
	elemVal, elemErr := s.modelProvider.ElemDependValue(rVal.Interface())
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

func (s *DeleteRunner) Delete() (err *cd.Result) {
	err = s.deleteHost()
	if err != nil {
		log.Errorf("Delete failed, s.deleteHost error:%s", err.Error())
		return
	}

	for _, field := range s.vModel.GetFields() {
		if field.IsBasic() {
			continue
		}

		err = s.deleteRelation(field, 0)
		if err != nil {
			log.Errorf("Delete failed, s.deleteRelation error:%s", err.Error())
			return
		}
	}

	return
}

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

	deleteRunner := NewDeleteRunner(vModel, s.executor, s.modelProvider, s.modelCodec, 0)
	err = deleteRunner.Delete()
	if err != nil {
		log.Errorf("Delete failed, s.deleteModel error:%s", err.Error())
		return
	}

	ret = vModel
	return
}
