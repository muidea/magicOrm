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

func (s *DeleteRunner) deleteHost(vModel model.Model) (err *cd.Error) {
	deleteResult, deleteErr := s.hBuilder.BuildDelete(vModel)
	if deleteErr != nil {
		err = deleteErr
		log.Errorf("deleteHost failed, s.hBuilder.BuildDelete error:%s", err.Error())
		return
	}

	_, err = s.executor.Execute(deleteResult.SQL(), deleteResult.Args()...)
	if err != nil {
		log.Errorf("deleteHost failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *DeleteRunner) deleteRelation(vModel model.Model, vField model.Field, deepLevel int) (err *cd.Error) {
	hostResult, relationResult, resultErr := s.hBuilder.BuildDeleteRelation(vModel, vField)
	if resultErr != nil {
		err = resultErr
		log.Errorf("deleteRelation failed, s.hBuilder.BuildDeleteRelation error:%s", err.Error())
		return
	}

	vType := vField.GetType()
	// 这里使用ElemType 是因为vFiled的指针表示该字段是否可选，
	elemType := vType.Elem()
	if !elemType.IsPtrType() {
		fieldErr := s.queryRelation(vModel, vField, maxDeepLevel-1)
		if fieldErr != nil {
			err = fieldErr
			log.Errorf("deleteRelation failed, s.queryRelation error:%s", err.Error())
			return
		}

		if model.IsStructType(vType.GetValue()) {
			err = s.deleteRelationSingleStructInner(vField, deepLevel)
			if err != nil {
				log.Errorf("deleteRelation failed, s.deleteRelationSingleStructInner error:%s", err.Error())
				return
			}
		} else if model.IsSliceType(vType.GetValue()) {
			err = s.deleteRelationSliceStructInner(vField, deepLevel)
			if err != nil {
				log.Errorf("deleteRelation failed, s.deleteRelationSliceStructInner error:%s", err.Error())
				return
			}
		}

		_, err = s.executor.Execute(hostResult.SQL(), hostResult.Args()...)
		if err != nil {
			log.Errorf("deleteRelation failed, s.executor.Execute error:%s", err.Error())
			return
		}
	}

	_, err = s.executor.Execute(relationResult.SQL(), relationResult.Args()...)
	if err != nil {
		log.Errorf("deleteRelation failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *DeleteRunner) deleteRelationSingleStructInner(vField model.Field, deepLevel int) (err *cd.Error) {
	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		err = rErr
		log.Errorf("deleteRelationSingleStructInner failed, s.modelProvider.GetTypeModel error:%s", err.Error())
		return
	}

	rModel, rErr = s.modelProvider.SetModelValue(rModel, vField.GetValue())
	if rErr != nil {
		err = rErr
		log.Errorf("deleteRelationSingleStructInner failed, s.modelProvider.SetModelValue error:%s", err.Error())
		return
	}

	rRunner := NewDeleteRunner(rModel, s.executor, s.modelProvider, s.modelCodec, deepLevel+1)
	err = rRunner.Delete()
	if err != nil {
		log.Errorf("deleteRelationSingleStructInner failed, rRunner.Delete error:%s", err.Error())
		return
	}

	return
}

func (s *DeleteRunner) deleteRelationSliceStructInner(vField model.Field, deepLevel int) (err *cd.Error) {
	sliceVal := vField.GetSliceValue()
	for _, val := range sliceVal {
		rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType().Elem())
		if rErr != nil {
			err = rErr
			log.Errorf("deleteRelationSliceStructInner failed, s.modelProvider.GetTypeModel error:%s", err.Error())
			return
		}

		rModel, rErr = s.modelProvider.SetModelValue(rModel, val)
		if rErr != nil {
			err = rErr
			log.Errorf("deleteRelationSliceStructInner failed, s.modelProvider.SetModelValue error:%s", err.Error())
			return
		}
		rRunner := NewDeleteRunner(rModel, s.executor, s.modelProvider, s.modelCodec, deepLevel+1)
		err = rRunner.Delete()
		if err != nil {
			log.Errorf("deleteRelationSliceStructInner failed, rRunner.Delete error:%s", err.Error())
			return
		}
	}

	return
}

func (s *DeleteRunner) Delete() (err *cd.Error) {
	err = s.deleteHost(s.vModel)
	if err != nil {
		log.Errorf("Delete failed, s.deleteHost error:%s", err.Error())
		return
	}

	for _, field := range s.vModel.GetFields() {
		if model.IsBasicField(field) {
			continue
		}

		err = s.deleteRelation(s.vModel, field, 0)
		if err != nil {
			log.Errorf("Delete failed, s.deleteRelation error:%s", err.Error())
			return
		}
	}

	return
}

func (s *impl) Delete(vModel model.Model) (ret model.Model, err *cd.Error) {
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
		log.Errorf("Delete failed, deleteRunner.Delete error:%s", err.Error())
		return
	}

	ret = vModel
	return
}
