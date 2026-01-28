package orm

import (
	"context"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

type DeleteRunner struct {
	baseRunner
	QueryRunner
}

func NewDeleteRunner(
	ctx context.Context,
	vModel models.Model,
	executor database.Executor,
	provider provider.Provider,
	modelCodec codec.Codec,
	deepLevel int) *DeleteRunner {
	baseRunner := newBaseRunner(ctx, vModel, executor, provider, modelCodec, false, deepLevel)
	return &DeleteRunner{
		baseRunner: baseRunner,
		QueryRunner: QueryRunner{
			baseRunner: baseRunner,
		},
	}
}

func (s *DeleteRunner) deleteHost(vModel models.Model) (err *cd.Error) {
	deleteResult, deleteErr := s.sqlBuilder.BuildDelete(vModel)
	if deleteErr != nil {
		err = deleteErr
		log.Errorf("deleteHost failed, s.sqlBuilder.BuildDelete error:%s", err.Error())
		return
	}

	_, err = s.executor.Execute(deleteResult.SQL(), deleteResult.Args()...)
	if err != nil {
		log.Errorf("deleteHost failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *DeleteRunner) deleteRelation(vModel models.Model, vField models.Field, deepLevel int) (err *cd.Error) {
	hostResult, relationResult, resultErr := s.sqlBuilder.BuildDeleteRelation(vModel, vField)
	if resultErr != nil {
		err = resultErr
		log.Errorf("deleteRelation failed, s.sqlBuilder.BuildDeleteRelation error:%s", err.Error())
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

		if models.IsStructType(vType.GetValue()) {
			err = s.deleteRelationSingleStructInner(vField, deepLevel)
			if err != nil {
				log.Errorf("deleteRelation failed, s.deleteRelationSingleStructInner error:%s", err.Error())
				return
			}
		} else if models.IsSliceType(vType.GetValue()) {
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

func (s *DeleteRunner) deleteRelationSingleStructInner(vField models.Field, deepLevel int) (err *cd.Error) {
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

	rRunner := NewDeleteRunner(s.context, rModel, s.executor, s.modelProvider, s.modelCodec, deepLevel+1)
	err = rRunner.Delete()
	if err != nil {
		log.Errorf("deleteRelationSingleStructInner failed, rRunner.Delete error:%s", err.Error())
		return
	}

	return
}

func (s *DeleteRunner) deleteRelationSliceStructInner(vField models.Field, deepLevel int) (err *cd.Error) {
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
		rRunner := NewDeleteRunner(s.context, rModel, s.executor, s.modelProvider, s.modelCodec, deepLevel+1)
		err = rRunner.Delete()
		if err != nil {
			log.Errorf("deleteRelationSliceStructInner failed, rRunner.Delete error:%s", err.Error())
			return
		}
	}

	return
}

func (s *DeleteRunner) Delete() (err *cd.Error) {
	if err = s.checkContext(); err != nil {
		return
	}

	err = s.deleteHost(s.vModel)
	if err != nil {
		log.Errorf("Delete failed, s.deleteHost error:%s", err.Error())
		return
	}

	for _, field := range s.vModel.GetFields() {
		if models.IsBasicField(field) {
			continue
		}

		err = s.deleteRelation(s.vModel, field, 0)
		if err != nil {
			log.Errorf("Delete relation field:%s failed, s.deleteRelation error:%s", field.GetName(), err.Error())
			return
		}
	}

	return
}

func (s *impl) Delete(vModel models.Model) (ret models.Model, err *cd.Error) {
	if err = s.CheckContext(); err != nil {
		return
	}

	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer s.finalTransaction(err)

	deleteRunner := NewDeleteRunner(s.context, vModel, s.executor, s.modelProvider, s.modelCodec, 0)
	err = deleteRunner.Delete()
	if err != nil {
		log.Errorf("Delete failed, deleteRunner.Delete error:%s", err.Error())
		return
	}

	ret = vModel
	return
}
