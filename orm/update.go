package orm

import (
	"context"
	"time"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/metrics"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/validation/errors"
	"log/slog"
)

func isReadOnlyField(field models.Field) bool {
	if field == nil {
		return false
	}

	spec := field.GetSpec()
	if spec == nil {
		return false
	}

	constraints := spec.GetConstraints()
	return constraints != nil && constraints.Has(models.KeyReadOnly)
}

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
		slog.Error("UpdateRunner updateHost BuildUpdate failed", "error", err.Error())
		return
	}

	_, err = s.executor.Execute(updateResult.SQL(), updateResult.Args()...)
	if err != nil {
		slog.Error("UpdateRunner updateHost Execute failed", "error", err.Error())
	}
	return
}

func (s *UpdateRunner) updateRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
	// 引用关系：单值指针（*T）或切片元素为指针（[]*T）；其余为包含关系
	isReference := (models.IsSliceField(vField) && vField.GetType().Elem().IsPtrType()) ||
		(!models.IsSliceField(vField) && models.IsPtrField(vField))
	if isReference {
		// 引用关系：只刷新关系（增删链接），不处理实体
		err = s.updateReferenceRelation(vModel, vField)
		if err != nil {
			slog.Error("UpdateRunner updateRelation updateReferenceRelation failed", "field", vField.GetName(), "error", err.Error())
		}
		return
	}
	// 包含关系：同步处理关系和实体（先 deleteRelation 再 insertRelation）
	err = s.updateContainRelation(vModel, vField)
	if err != nil {
		slog.Error("UpdateRunner updateRelation updateContainRelation failed", "field", vField.GetName(), "error", err.Error())
	}
	return
}

func (s *UpdateRunner) Update() (ret models.Model, err *cd.Error) {
	if err = s.checkContext(); err != nil {
		return
	}

	err = s.updateHost(s.vModel)
	if err != nil {
		slog.Error("UpdateRunner Update updateHost failed", "error", err.Error())
		return
	}

	for _, field := range s.vModel.GetFields() {
		// 忽略基础字段和未赋值的字段
		// 未赋值则认为不需要对该字段进行更新
		if models.IsBasicField(field) || !models.IsAssignedField(field) || isReadOnlyField(field) {
			continue
		}

		err = s.updateRelation(s.vModel, field)
		if err != nil {
			slog.Error("UpdateRunner Update updateRelation failed", "field", field.GetName(), "error", err.Error())
			return
		}
	}

	ret = s.vModel
	return
}

func (s *impl) Update(vModel models.Model) (ret models.Model, err *cd.Error) {
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		if ormMetricCollector != nil {
			ormMetricCollector.RecordOperation(string(metrics.OperationUpdate), vModel, duration, err)
		}
	}()

	if err = s.CheckContext(); err != nil {
		return
	}

	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	// Validate model before update
	validationErr := s.validateModel(vModel, errors.ScenarioUpdate)
	if validationErr != nil {
		err = validationErr
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
		slog.Error("Update UpdateRunner.Update failed", "pkgKey", vModel.GetPkgKey(), "error", err.Error())
		return
	}

	return
}
