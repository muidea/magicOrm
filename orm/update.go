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

	hostUpdateDuration     time.Duration
	relationUpdateDuration time.Duration
	relationUpdateCount    int
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
	startTime := time.Now()
	defer func() {
		s.hostUpdateDuration += time.Since(startTime)
	}()

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
	startTime := time.Now()
	defer func() {
		s.relationUpdateDuration += time.Since(startTime)
		if err == nil {
			s.relationUpdateCount++
		}
	}()

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

func assignedReadOnlyBasicFields(vModel models.Model) (ret []string) {
	if vModel == nil {
		return nil
	}

	for _, field := range vModel.GetFields() {
		if field == nil || models.IsPrimaryField(field) || !models.IsBasicField(field) || !models.IsAssignedField(field) || !isReadOnlyField(field) {
			continue
		}
		ret = append(ret, field.GetName())
	}

	return ret
}

func buildPrimaryQueryModel(vModel models.Model) (ret models.Model, err *cd.Error) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "query model is nil")
		return
	}

	ret = vModel.Copy(models.MetaView)
	ret.Reset()

	pkField := vModel.GetPrimaryField()
	if pkField == nil || !pkField.GetValue().IsValid() {
		err = cd.NewError(cd.IllegalParam, "primary key is invalid")
		return
	}

	err = ret.SetPrimaryFieldValue(pkField.GetValue().Get())
	return
}

func (s *UpdateRunner) queryStoredModelForReadOnlyFields(vModel models.Model) (ret models.Model, err *cd.Error) {
	queryModel, queryErr := buildPrimaryQueryModel(vModel)
	if queryErr != nil {
		err = queryErr
		return
	}

	filter, filterErr := getModelFilter(queryModel, s.modelProvider, s.modelCodec)
	if filterErr != nil {
		err = filterErr
		return
	}

	responseModel := queryModel.Copy(models.DetailView)
	queryMask, maskErr := buildFullQueryMaskModel(responseModel)
	if maskErr != nil {
		err = maskErr
		return
	}

	queryRunner := NewQueryRunner(s.context, queryMask, responseModel, false, s.executor, s.modelProvider, s.modelCodec, false, 0)
	modelList, queryExecErr := queryRunner.Query(filter)
	if queryExecErr != nil {
		err = queryExecErr
		return
	}
	if len(modelList) == 0 {
		err = cd.NewError(cd.NotFound, "stored model not found")
		return
	}

	ret = modelList[0]
	return
}

func restoreReadOnlyBasicFields(dst models.Model, src models.Model, fieldNames []string) {
	if dst == nil || src == nil {
		return
	}

	for _, fieldName := range fieldNames {
		srcField := src.GetField(fieldName)
		if srcField == nil || !srcField.GetValue().IsValid() {
			continue
		}
		_ = dst.SetFieldValue(fieldName, srcField.GetValue().Get())
	}
}

func (s *UpdateRunner) Update() (ret models.Model, err *cd.Error) {
	if err = s.checkContext(); err != nil {
		return
	}

	readOnlyFieldNames := assignedReadOnlyBasicFields(s.vModel)
	var storedModel models.Model
	if len(readOnlyFieldNames) > 0 {
		storedModel, err = s.queryStoredModelForReadOnlyFields(s.vModel)
		if err != nil {
			slog.Warn("UpdateRunner prequery readonly fields failed", "pkgKey", s.vModel.GetPkgKey(), "error", err.Error())
			err = nil
		}
	}

	if hasAssignedWritableBasicFields(s.vModel) {
		err = s.updateHost(s.vModel)
		if err != nil {
			slog.Error("UpdateRunner Update updateHost failed", "error", err.Error())
			return
		}
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

	if storedModel != nil {
		restoreReadOnlyBasicFields(s.vModel, storedModel, readOnlyFieldNames)
	}

	ret, err = projectWriteResponseModel(s.vModel, s.modelProvider)
	return
}

func hasAssignedWritableBasicFields(vModel models.Model) bool {
	if vModel == nil {
		return false
	}

	for _, field := range vModel.GetFields() {
		if !models.IsBasicField(field) || models.IsPrimaryField(field) || !models.IsAssignedField(field) || isReadOnlyField(field) {
			continue
		}
		return true
	}

	return false
}

func updateRequiresTransaction(vModel models.Model) bool {
	if vModel == nil {
		return false
	}

	for _, field := range vModel.GetFields() {
		if models.IsBasicField(field) || !models.IsAssignedField(field) || isReadOnlyField(field) {
			continue
		}
		return true
	}

	return false
}

func (s *impl) Update(vModel models.Model) (ret models.Model, err *cd.Error) {
	startTime := time.Now()
	var validationDuration time.Duration
	var txBeginDuration time.Duration
	var txFinalizeDuration time.Duration
	usedTransaction := false

	defer func() {
		duration := time.Since(startTime)
		if ormMetricCollector != nil {
			ormMetricCollector.RecordOperation(string(metrics.OperationUpdate), vModel, duration, cd.ToStdError(err))
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
	validationStart := time.Now()
	validationErr := s.validateModel(vModel, errors.ScenarioUpdate)
	validationDuration = time.Since(validationStart)
	if validationErr != nil {
		err = validationErr
		return
	}

	// Pure host-field updates compile down to a single UPDATE statement. Keeping
	// them out of an explicit transaction avoids an extra begin/commit round-trip
	// on the write hot path; relation updates still need transactional wrapping.
	if updateRequiresTransaction(vModel) {
		usedTransaction = true
		txBeginStart := time.Now()
		err = s.executor.BeginTransaction()
		txBeginDuration = time.Since(txBeginStart)
		if err != nil {
			return
		}
		defer func() {
			txFinalizeStart := time.Now()
			s.finalTransaction(err)
			txFinalizeDuration = time.Since(txFinalizeStart)
		}()
	}

	updateRunner := NewUpdateRunner(s.context, vModel, s.executor, s.modelProvider, s.modelCodec)
	ret, err = updateRunner.Update()
	if err != nil {
		slog.Error("Update UpdateRunner.Update failed", "pkgKey", vModel.GetPkgKey(), "error", err.Error())
		return
	}

	totalDuration := time.Since(startTime)
	if totalDuration >= 10*time.Millisecond {
		slog.Info("UpdateRunner profile",
			"pkgKey", vModel.GetPkgKey(),
			"total_ms", totalDuration.Seconds()*1000,
			"validation_ms", validationDuration.Seconds()*1000,
			"tx_begin_ms", txBeginDuration.Seconds()*1000,
			"tx_finalize_ms", txFinalizeDuration.Seconds()*1000,
			"host_update_ms", updateRunner.hostUpdateDuration.Seconds()*1000,
			"relation_update_ms", updateRunner.relationUpdateDuration.Seconds()*1000,
			"relation_update_count", updateRunner.relationUpdateCount,
			"used_tx", usedTransaction,
			"has_basic_update", hasAssignedWritableBasicFields(vModel),
		)
	}

	return
}
