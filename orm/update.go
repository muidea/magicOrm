package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

type UpdateRunner struct {
	baseRunner
	QueryRunner
	InsertRunner
	DeleteRunner
}

func NewUpdateRunner(
	vModel model.Model,
	executor executor.Executor,
	provider provider.Provider,
	modelCodec codec.Codec) *UpdateRunner {
	baseRunner := newBaseRunner(vModel, executor, provider, modelCodec, false, 0)
	return &UpdateRunner{
		baseRunner: baseRunner,
		QueryRunner: QueryRunner{
			baseRunner: baseRunner,
		},
		InsertRunner: InsertRunner{
			baseRunner: baseRunner,
			QueryRunner: QueryRunner{
				baseRunner: baseRunner,
			},
		},
		DeleteRunner: DeleteRunner{
			baseRunner: baseRunner,
			QueryRunner: QueryRunner{
				baseRunner: baseRunner,
			},
		},
	}
}

func (s *UpdateRunner) updateHost() (err *cd.Result) {
	updateResult, updateErr := s.hBuilder.BuildUpdate(s.baseRunner.vModel)
	if updateErr != nil {
		err = updateErr
		log.Errorf("updateHost failed, builderVal.BuildUpdate error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(updateResult.SQL(), updateResult.Args()...)
	if err != nil {
		log.Errorf("updateHost failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *UpdateRunner) updateRelation(vField model.Field, rModel model.Model) (err *cd.Result) {
	err = s.deleteRelation(vField, rModel, 0)
	if err != nil {
		log.Errorf("updateRelation failed, s.deleteRelation error:%s", err.Error())
		return
	}

	err = s.insertRelation(vField, rModel)
	if err != nil {
		log.Errorf("updateRelation failed, s.insertRelation error:%s", err.Error())
	}
	return
}

func (s *UpdateRunner) Update() (ret model.Model, err *cd.Result) {
	err = s.updateHost()
	if err != nil {
		log.Errorf("Update failed, s.updateSingle error:%s", err.Error())
		return
	}

	for _, field := range s.vModel.GetFields() {
		if field.IsBasic() || !field.GetValue().IsValid() {
			continue
		}

		rModel, rErr := s.modelProvider.GetTypeModel(field.GetType())
		if rErr != nil {
			err = rErr
			log.Errorf("Update failed, s.modelProvider.GetTypeModel error:%s", err.Error())
			return
		}

		err = s.updateRelation(field, rModel)
		if err != nil {
			log.Errorf("Update failed, s.updateRelation error:%s", err.Error())
			return
		}
	}

	ret = s.vModel
	return
}

func (s *impl) Update(vModel model.Model) (ret model.Model, err *cd.Result) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer s.finalTransaction(err)

	updateRunner := NewUpdateRunner(vModel, s.executor, s.modelProvider, s.modelCodec)
	ret, err = updateRunner.Update()
	if err != nil {
		log.Errorf("Update failed, updateRunner.Update() error:%s", err.Error())
		return
	}

	return
}
