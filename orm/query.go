package orm

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/metrics"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

type resultItems []any
type resultItemsList []resultItems

type QueryRunner struct {
	baseRunner
	responseModel  models.Model
	responseByMask bool
}

func buildFullQueryMaskModel(vModel models.Model) (ret models.Model, err *cd.Error) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "query model is nil")
		return
	}

	maskModel := vModel.Copy(models.OriginView)
	for _, field := range maskModel.GetFields() {
		if !models.IsBasicField(field) || models.IsValidField(field) || models.IsPtrField(field) {
			continue
		}

		initValue, initErr := field.GetType().Interface(nil)
		if initErr != nil {
			err = initErr
			return
		}
		if initValue == nil {
			continue
		}

		setErr := field.SetValue(initValue.Get())
		if setErr != nil {
			err = setErr
			return
		}
	}

	ret = maskModel
	return
}

type queryResponseMaskProvider interface {
	HasValueMask() bool
	ResponseModel() models.Model
}

type explicitResponseModelProvider interface {
	ExplicitResponseModel() models.Model
}

type responseFieldChecker interface {
	ResponseIncludesField(name string) bool
}

func buildQueryResponseModel(vModel models.Model, filter models.Filter) (ret models.Model, responseByMask bool, err *cd.Error) {
	if filter != nil {
		if responseProvider, ok := filter.(queryResponseMaskProvider); ok {
			if responseProvider.HasValueMask() {
				responseByMask = true
				if explicitProvider, explicitOK := filter.(explicitResponseModelProvider); explicitOK {
					ret = explicitProvider.ExplicitResponseModel()
				} else {
					ret = filter.MaskModel()
				}
				return
			}

			responseModel := responseProvider.ResponseModel()
			if responseModel != nil {
				ret = responseModel
				return
			}
		}

		ret = filter.MaskModel()
		return
	}

	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "query response model is nil")
		return
	}

	ret = vModel.Copy(models.OriginView)
	return
}

func fieldIncludedInResponse(responseModel models.Model, field models.Field, responseByMask bool) bool {
	if field == nil {
		return false
	}
	if models.IsPrimaryField(field) {
		return true
	}
	if !responseByMask {
		if fieldChecker, ok := responseModel.(responseFieldChecker); ok {
			return fieldChecker.ResponseIncludesField(field.GetName())
		}
	}
	return models.IsValidField(field) || models.IsAssignedField(field)
}

func applyQueryResponseModel(vModel, responseModel models.Model, responseByMask bool) models.Model {
	if vModel == nil || responseModel == nil {
		return vModel
	}

	projectedModel := responseModel.Copy(models.OriginView)
	primaryField := vModel.GetPrimaryField()
	if primaryField != nil && (models.IsValidField(primaryField) || models.IsAssignedField(primaryField)) {
		_ = projectedModel.SetPrimaryFieldValue(primaryField.GetValue().Get())
	}

	for _, field := range projectedModel.GetFields() {
		if !fieldIncludedInResponse(responseModel, field, responseByMask) {
			continue
		}

		if models.IsPrimaryField(field) {
			continue
		}

		sourceField := vModel.GetField(field.GetName())
		if sourceField == nil {
			continue
		}
		if !models.IsValidField(sourceField) && !models.IsAssignedField(sourceField) {
			_ = projectedModel.SetFieldValue(field.GetName(), nil)
			continue
		}

		_ = projectedModel.SetFieldValue(field.GetName(), sourceField.GetValue().Get())
	}

	return projectedModel
}

func NewQueryRunner(
	ctx context.Context,
	vModel models.Model,
	responseModel models.Model,
	responseByMask bool,
	executor database.Executor,
	provider provider.Provider,
	modelCodec codec.Codec,
	batchFilter bool,
	deepLevel int) *QueryRunner {

	return &QueryRunner{
		baseRunner:     newBaseRunner(ctx, vModel, executor, provider, modelCodec, batchFilter, deepLevel),
		responseModel:  responseModel,
		responseByMask: responseByMask,
	}
}

func (s *QueryRunner) innerQuery(vModel models.Model, filter models.Filter) (ret resultItemsList, err *cd.Error) {
	queryResult, queryErr := s.sqlBuilder.BuildQuery(vModel, filter)
	if queryErr != nil {
		err = queryErr
		slog.Error("QueryRunner innerQuery BuildQuery failed", "error", err.Error())
		return
	}

	_, err = s.executor.Query(queryResult.SQL(), false, queryResult.Args()...)
	if err != nil {
		slog.Error("QueryRunner innerQuery executor.Query failed", "error", err.Error())
		return
	}
	defer s.executor.Finish()

	queryList := resultItemsList{}
	for s.executor.Next() {
		itemValues, itemErr := s.sqlBuilder.BuildModuleValueHolder(vModel)
		if itemErr != nil {
			err = itemErr
			slog.Error("QueryRunner innerQuery BuildModuleValueHolder failed", "error", err.Error())
			return
		}
		referenceVal := make([]any, len(itemValues))
		for idx := range itemValues {
			referenceVal[idx] = &itemValues[idx]
		}

		err = s.executor.GetField(referenceVal...)
		if err != nil {
			slog.Error("QueryRunner failed", "error", err.Error())
			return
		}

		queryList = append(queryList, itemValues)
	}

	ret = queryList
	return
}

func (s *QueryRunner) innerAssign(vModel models.Model, queryVal resultItems, deepLevel int) (ret models.Model, err *cd.Error) {
	offset := 0
	qModel := vModel.Copy(models.OriginView)
	for _, field := range qModel.GetFields() {
		// 只处理基础字段；与 builder 一致：已赋值或值类型 slice 才参与 SELECT，故只对这些字段赋值
		if !models.IsBasicField(field) {
			continue
		}
		if !models.IsValidField(field) && !(models.IsSliceField(field) && !models.IsPtrField(field)) {
			continue
		}
		// 检查 wo 约束，这些字段在查询时应该被排除
		fSpec := field.GetSpec()
		constraints := fSpec.GetConstraints()
		if constraints != nil && constraints.Has(models.KeyWriteOnly) {
			continue
		}

		err = s.assignBasicField(field, queryVal[offset])
		if err != nil {
			slog.Error("QueryRunner failed", "error", err.Error())
			return
		}
		offset++
	}

	for _, field := range qModel.GetFields() {
		// 只对关系字段（非基础）加载关系；关系字段即使当前为 nil 也从 DB 加载，与「nil=未赋值、[]=已赋值」一致
		if models.IsBasicField(field) {
			continue
		}
		err = s.assignModelField(qModel, field, deepLevel)
		if err != nil {
			slog.Error("QueryRunner failed", "error", err.Error())
			return
		}
	}

	ret = qModel
	return
}

func (s *QueryRunner) assignModelField(vModel models.Model, vField models.Field, deepLevel int) (err *cd.Error) {
	vErr := s.queryRelation(vModel, vField, deepLevel)
	if vErr != nil {
		err = vErr
		slog.Error("QueryRunner failed", "error", err.Error())
		return
	}

	return
}

func (s *QueryRunner) assignBasicField(vField models.Field, val any) (err *cd.Error) {
	if val == nil {
		return
	}

	fVal, fErr := s.modelCodec.ExtractBasicFieldValue(vField, val)
	if fErr != nil {
		err = fErr
		slog.Error("QueryRunner failed", "error", err.Error())
		return
	}

	vField.SetValue(fVal)
	return
}

func (s *QueryRunner) queryRelation(vModel models.Model, vField models.Field, deepLevel int) (err *cd.Error) {
	if deepLevel > maxDeepLevel {
		return
	}

	if models.IsSliceField(vField) {
		err = s.querySliceRelation(vModel, vField, deepLevel)
		if err != nil {
			slog.Error("QueryRunner failed", "error", err.Error())
		}
		return
	}

	err = s.querySingleRelation(vModel, vField, deepLevel)
	if err != nil {
		slog.Error("QueryRunner failed", "error", err.Error())
	}
	return
}

func (s *QueryRunner) querySingleRelation(vModel models.Model, vField models.Field, deepLevel int) (err *cd.Error) {
	valueList, valueErr := s.innerQueryRelationKeys(vModel, vField)
	if valueErr != nil {
		err = valueErr
		slog.Error("QueryRunner failed", "error", err.Error())
		return
	}

	vType := vField.GetType()
	valueSize := len(valueList)
	if valueSize == 0 {
		if vType.IsPtrType() {
			return
		}
		slog.Warn("query relation failed", "field", vField.GetName())
		return
	}

	rvErr := s.innerQueryRelationSingleModel(valueList[0], vField, deepLevel)
	if rvErr != nil {
		slog.Error("QueryRunner failed", "error", rvErr.Error())
		err = rvErr
		return
	}
	return
}

func (s *QueryRunner) querySliceRelation(vModel models.Model, vField models.Field, deepLevel int) (err *cd.Error) {
	valueList, valueErr := s.innerQueryRelationKeys(vModel, vField)
	if valueErr != nil {
		err = valueErr
		slog.Error("QueryRunner failed", "error", err.Error())
		return
	}
	valueSize := len(valueList)
	if valueSize == 0 {
		return
	}

	rModelErr := s.innerQueryRelationSliceModel(valueList, vField, deepLevel)
	if rModelErr != nil {
		err = rModelErr
		slog.Error("QueryRunner failed", "error", rModelErr.Error())
		return
	}
	return
}

func (s *QueryRunner) innerQueryRelationKeys(vModel models.Model, vField models.Field) (ret resultItems, err *cd.Error) {
	relationResult, relationErr := s.sqlBuilder.BuildQueryRelation(vModel, vField)
	if relationErr != nil {
		err = relationErr
		slog.Error("QueryRunner failed", "error", err.Error())
		return
	}

	values := resultItems{}
	func() {
		_, err = s.executor.Query(relationResult.SQL(), false, relationResult.Args()...)
		if err != nil {
			slog.Error("QueryRunner failed", "error", err.Error())
			return
		}
		defer s.executor.Finish()

		for s.executor.Next() {
			var idVal any
			err = s.executor.GetField(&idVal)
			if err != nil {
				slog.Error("QueryRunner failed", "error", err.Error())
				return
			}
			values = append(values, idVal)
		}
	}()

	if err != nil {
		return
	}

	ret = values
	return
}

func (s *QueryRunner) innerQueryRelationSingleModel(id any, vField models.Field, deepLevel int) (err *cd.Error) {
	vField.Reset()
	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		err = rErr
		slog.Error("QueryRunner failed", "error", err.Error())
		return
	}

	rVal, rErr := s.modelCodec.ExtractBasicFieldValue(rModel.GetPrimaryField(), id)
	if rErr != nil {
		err = rErr
		slog.Error("QueryRunner failed", "error", err.Error())
		return
	}
	rModel.SetPrimaryFieldValue(rVal)
	vFilter, vErr := getModelFilter(rModel, s.modelProvider, s.modelCodec)
	if vErr != nil {
		err = vErr
		slog.Error("QueryRunner assignBasicField failed", "fieldId", id, "error", err.Error())
		return
	}

	queryMask, maskErr := buildFullQueryMaskModel(rModel)
	if maskErr != nil {
		err = maskErr
		slog.Error("QueryRunner buildFullQueryMaskModel failed", "fieldId", id, "error", err.Error())
		return
	}

	rQueryRunner := NewQueryRunner(s.context, queryMask, queryMask, false, s.executor, s.modelProvider, s.modelCodec, false, deepLevel+1)
	queryVal, queryErr := rQueryRunner.Query(vFilter)
	if queryErr != nil {
		err = queryErr
		slog.Error("QueryRunner assignBasicField failed", "fieldId", id, "error", err.Error())
		return
	}
	if len(queryVal) > 1 {
		errMsg := fmt.Sprintf("match more than one model, model:%s, id:%v", rModel.GetPkgKey(), id)
		slog.Warn("innerQueryRelationSingleModel failed", "error", errMsg)
		err = cd.NewError(cd.Unexpected, errMsg)
		return
	}

	if len(queryVal) > 0 {
		vField.SetValue(queryVal[0].Interface(vField.GetType().Elem().IsPtrType()))
		return
	}

	if deepLevel < maxDeepLevel {
		// 到这里说明未查询到数据，说明存在数据表之间数据不一致
		// 这种情况下直接返回nil，后续要考虑进行脏数据检测
		slog.Warn("query relation failed, miss relation data", "model", rModel.GetPkgKey(), "id", id)
	}
	return
}

func (s *QueryRunner) innerQueryRelationSliceModel(ids []any, vField models.Field, deepLevel int) (err *cd.Error) {
	// 这里主动重置，避免VFiled的旧数据干扰
	vField.Reset()
	for _, id := range ids {
		svModel, svErr := s.modelProvider.GetTypeModel(vField.GetType().Elem())
		if svErr != nil {
			err = svErr
			slog.Error("QueryRunner failed", "error", err.Error())
			return
		}
		rVal, rErr := s.modelCodec.ExtractBasicFieldValue(svModel.GetPrimaryField(), id)
		if rErr != nil {
			err = rErr
			slog.Error("QueryRunner failed", "error", err.Error())
			return
		}
		svModel.SetPrimaryFieldValue(rVal)
		vFilter, vErr := getModelFilter(svModel, s.modelProvider, s.modelCodec)
		if vErr != nil {
			err = vErr
			slog.Error("QueryRunner failed", "error", err.Error())
			return
		}

		queryMask, maskErr := buildFullQueryMaskModel(svModel)
		if maskErr != nil {
			err = maskErr
			slog.Error("QueryRunner buildFullQueryMaskModel failed", "error", err.Error())
			return
		}

		rQueryRunner := NewQueryRunner(s.context, queryMask, queryMask, false, s.executor, s.modelProvider, s.modelCodec, false, deepLevel+1)
		queryVal, queryErr := rQueryRunner.Query(vFilter)
		if queryErr != nil {
			err = queryErr
			slog.Error("QueryRunner failed", "error", err.Error())
			return
		}

		if len(queryVal) > 0 {
			vField.AppendSliceValue(queryVal[0].Interface(vField.GetType().Elem().IsPtrType()))
		}
	}

	return
}

func (s *QueryRunner) Query(filter models.Filter) (ret []models.Model, err *cd.Error) {
	if err = s.checkContext(); err != nil {
		return
	}

	queryValueList, queryValueErr := s.innerQuery(s.vModel, filter)
	if queryValueErr != nil {
		err = queryValueErr
		slog.Error("QueryRunner failed", "error", err.Error())
		return
	}

	queryCount := len(queryValueList)
	if queryCount == 0 {
		return
	}
	if !s.batchFilter && queryCount > 1 {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("matched model:%s %d items value", s.vModel.GetPkgKey(), queryCount))
		slog.Warn("Query failed", "error", err.Error())
		return
	}

	sliceValue := []models.Model{}
	for idx := range queryValueList {
		modelVal, modelErr := s.innerAssign(s.vModel, queryValueList[idx], 0)
		if modelErr != nil {
			err = modelErr
			slog.Error("QueryRunner failed", "error", err.Error())
			return
		}
		modelVal = applyQueryResponseModel(modelVal, s.responseModel, s.responseByMask)

		sliceValue = append(sliceValue, modelVal)
	}

	ret = sliceValue
	return
}

func (s *impl) Query(vModel models.Model) (ret models.Model, err *cd.Error) {
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		if ormMetricCollector != nil {
			ormMetricCollector.RecordOperation(string(metrics.OperationQuery), vModel, duration, err)
		}
	}()

	if err = s.CheckContext(); err != nil {
		return
	}

	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "query model is nil")
		return
	}

	// 这里主动Copy一份出来，是为了避免在查询数据过程中对源数据产生了干扰
	vModel = vModel.Copy(models.OriginView)
	vFilter, vErr := getModelFilter(vModel, s.modelProvider, s.modelCodec)
	if vErr != nil {
		err = vErr
		slog.Error("Query getModelFilter failed", "pkgKey", vModel.GetPkgKey(), "error", err.Error())
		return
	}

	responseModel, responseByMask, responseErr := buildQueryResponseModel(vModel, nil)
	if responseErr != nil {
		err = responseErr
		slog.Error("Query buildQueryResponseModel failed", "pkgKey", vModel.GetPkgKey(), "error", err.Error())
		return
	}

	queryMask, maskErr := buildFullQueryMaskModel(responseModel)
	if maskErr != nil {
		err = maskErr
		slog.Error("Query buildFullQueryMaskModel failed", "pkgKey", vModel.GetPkgKey(), "error", err.Error())
		return
	}

	vQueryRunner := NewQueryRunner(s.context, queryMask, responseModel, responseByMask, s.executor, s.modelProvider, s.modelCodec, false, 0)
	queryVal, queryErr := vQueryRunner.Query(vFilter)
	if queryErr != nil {
		err = queryErr
		slog.Error("Query QueryRunner.Query failed", "pkgKey", vModel.GetPkgKey(), "error", err.Error())
		return
	}
	if len(queryVal) != 0 {
		ret = queryVal[0]
		return
	}

	err = cd.NewError(cd.NotFound, fmt.Sprintf("no records found matching the model criteria, model pkgKey: %s, filter: %v", vModel.GetPkgKey(), vFilter))
	return
}
