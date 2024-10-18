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
	deleteResult, deleteErr := s.hBuilder.BuildDelete(s.vModel)
	if deleteErr != nil {
		err = deleteErr
		log.Errorf("deleteHost failed, s.hBuilder.BuildDelete error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(deleteResult.SQL(), deleteResult.Args()...)
	if err != nil {
		log.Errorf("deleteHost failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *DeleteRunner) deleteRelation(vField model.Field, rModel model.Model, deepLevel int) (err *cd.Result) {
	hostResult, relationResult, resultErr := s.hBuilder.BuildDeleteRelation(s.vModel, vField, rModel)
	if resultErr != nil {
		err = resultErr
		log.Errorf("deleteRelation failed, s.hBuilder.BuildDeleteRelation error:%s", err.Error())
		return
	}

	vType := vField.GetType()
	elemType := vType.Elem()
	if !elemType.IsPtrType() {
		fieldVal, fieldErr := s.queryRelation(s.vModel, vField, rModel, maxDeepLevel-1)
		if fieldErr != nil {
			err = fieldErr
			log.Errorf("deleteRelation failed, s.queryRelation error:%s", err.Error())
			return
		}

		if model.IsStructType(vType.GetValue()) {
			err = s.deleteRelationSingleStructInner(fieldVal, rModel, deepLevel)
			if err != nil {
				log.Errorf("deleteRelation failed, s.deleteRelationSingleStructInner error:%s", err.Error())
				return
			}
		} else if model.IsSliceType(vType.GetValue()) {
			err = s.deleteRelationSliceStructInner(fieldVal, rModel, deepLevel)
			if err != nil {
				log.Errorf("deleteRelation failed, s.deleteRelationSliceStructInner error:%s", err.Error())
				return
			}
		}

		_, _, err = s.executor.Execute(hostResult.SQL(), hostResult.Args()...)
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

func (s *DeleteRunner) deleteRelationSingleStructInner(rVal model.Value, rModel model.Model, deepLevel int) (err *cd.Result) {
	relationModel, relationErr := s.modelProvider.SetModelValue(rModel, rVal)
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

func (s *DeleteRunner) deleteRelationSliceStructInner(rVal model.Value, rModel model.Model, deepLevel int) (err *cd.Result) {
	elemVal, elemErr := s.modelProvider.ElemDependValue(rVal.Interface())
	if elemErr != nil {
		err = elemErr
		log.Errorf("deleteRelationSliceStructInner failed, s.modelProvider.ElemDependValue error:%s", err.Error())
		return
	}

	for idx := 0; idx < len(elemVal); idx++ {
		err = s.deleteRelationSingleStructInner(elemVal[idx], rModel.Copy(true), deepLevel)
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

		rModel, rErr := s.modelProvider.GetTypeModel(field.GetType())
		if rErr != nil {
			err = rErr
			log.Errorf("Delete failed, s.modelProvider.GetTypeModel error:%s", err.Error())
			return
		}
		err = s.deleteRelation(field, rModel, 0)
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
