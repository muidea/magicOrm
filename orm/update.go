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

type UpdateRunner struct {
	baseRunner
	QueryRunner
	InsertRunner
	DeleteRunner
}

func NewUpdateRunner(
	ctx context.Context,
	vModel models.Model,
	executor database.Executor,
	provider provider.Provider,
	modelCodec codec.Codec) *UpdateRunner {
	baseRunner := newBaseRunner(ctx, vModel, executor, provider, modelCodec, false, 0)
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

func (s *UpdateRunner) updateHost(vModel models.Model) (err *cd.Error) {
	updateResult, updateErr := s.sqlBuilder.BuildUpdate(vModel)
	if updateErr != nil {
		err = updateErr
		log.Errorf("updateHost failed, s.sqlBuilder.BuildUpdate error:%s", err.Error())
		return
	}

	_, err = s.executor.Execute(updateResult.SQL(), updateResult.Args()...)
	if err != nil {
		log.Errorf("updateHost failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *UpdateRunner) updateRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
	newVal := vField.GetValue().Get()
	err = s.deleteRelation(vModel, vField, 0)
	if err != nil {
		log.Errorf("updateRelation failed, s.deleteRelation error:%s", err.Error())
		return
	}
	// TODO 这里最合理的逻辑应该是先查询出当前值，与新值进行差异比较
	// 再根据比较后的结果进行处理
	// 目前先粗暴点，直接删除再插入
	vField.SetValue(newVal)
	err = s.insertRelation(vModel, vField)
	if err != nil {
		log.Errorf("updateRelation failed, s.insertRelation error:%s", err.Error())
	}
	return
}

func (s *UpdateRunner) Update() (ret models.Model, err *cd.Error) {
	err = s.updateHost(s.vModel)
	if err != nil {
		log.Errorf("Update failed, s.updateSingle error:%s", err.Error())
		return
	}

	for _, field := range s.vModel.GetFields() {
		if models.IsBasicField(field) || !models.IsValidField(field) {
			continue
		}

		err = s.updateRelation(s.vModel, field)
		if err != nil {
			log.Errorf("Update failed, s.updateRelation error:%s", err.Error())
			return
		}
	}

	ret = s.vModel
	return
}

func (s *impl) Update(vModel models.Model) (ret models.Model, err *cd.Error) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer s.finalTransaction(err)

	updateRunner := NewUpdateRunner(s.context, vModel, s.executor, s.modelProvider, s.modelCodec)
	ret, err = updateRunner.Update()
	if err != nil {
		log.Errorf("Update failed, updateRunner.Update() error:%s", err.Error())
		return
	}

	return
}
