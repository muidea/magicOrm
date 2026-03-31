package orm

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
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
	relationCache  map[string]models.Model
	relationMisses map[string]struct{}
	relationEdges  map[string][]any
}

type relationPrefetchGroup struct {
	model   models.Model
	field   models.Field
	leftIDs []any
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
		relationCache:  map[string]models.Model{},
		relationMisses: map[string]struct{}{},
		relationEdges:  map[string][]any{},
	}
}

func relationCacheKey(pkgKey string, id any) string {
	return fmt.Sprintf("%s:%v", pkgKey, id)
}

func relationEdgeKey(pkgKey, fieldName string, id any) string {
	return fmt.Sprintf("%s:%s:%v", pkgKey, fieldName, id)
}

func buildTypedValueSlice(vals []any) any {
	if len(vals) == 0 {
		return vals
	}

	elemType := reflect.TypeOf(vals[0])
	if elemType == nil {
		return vals
	}

	sliceVal := reflect.MakeSlice(reflect.SliceOf(elemType), 0, len(vals))
	for _, val := range vals {
		if reflect.TypeOf(val) != elemType {
			return vals
		}
		sliceVal = reflect.Append(sliceVal, reflect.ValueOf(val))
	}

	return sliceVal.Interface()
}

func (s *QueryRunner) getCachedRelationModel(pkgKey string, id any) models.Model {
	if s == nil || s.relationCache == nil {
		return nil
	}

	return s.relationCache[relationCacheKey(pkgKey, id)]
}

func (s *QueryRunner) isRelationMiss(pkgKey string, id any) bool {
	if s == nil || s.relationMisses == nil {
		return false
	}

	_, ok := s.relationMisses[relationCacheKey(pkgKey, id)]
	return ok
}

func (s *QueryRunner) cacheRelationMiss(pkgKey string, id any) {
	if s == nil || s.relationMisses == nil {
		return
	}

	s.relationMisses[relationCacheKey(pkgKey, id)] = struct{}{}
}

func (s *QueryRunner) clearRelationMiss(pkgKey string, id any) {
	if s == nil || s.relationMisses == nil {
		return
	}

	delete(s.relationMisses, relationCacheKey(pkgKey, id))
}

func (s *QueryRunner) cacheRelationModel(model models.Model) {
	if s == nil || s.relationCache == nil || model == nil {
		return
	}

	pkField := model.GetPrimaryField()
	if pkField == nil || !models.IsAssignedField(pkField) {
		return
	}

	pkVal := pkField.GetValue().Get()
	s.relationCache[relationCacheKey(model.GetPkgKey(), pkVal)] = model
	s.clearRelationMiss(model.GetPkgKey(), pkVal)
}

func (s *QueryRunner) cacheRelationEdge(pkgKey, fieldName string, leftID any, rightIDs []any) {
	if s == nil || s.relationEdges == nil {
		return
	}

	edgeIDs := make([]any, len(rightIDs))
	copy(edgeIDs, rightIDs)
	s.relationEdges[relationEdgeKey(pkgKey, fieldName, leftID)] = edgeIDs
}

func (s *QueryRunner) getCachedRelationEdge(pkgKey, fieldName string, leftID any) (ret []any, ok bool) {
	if s == nil || s.relationEdges == nil {
		return
	}

	ret, ok = s.relationEdges[relationEdgeKey(pkgKey, fieldName, leftID)]
	return
}

func (s *QueryRunner) shouldLoadRelationField(field models.Field) bool {
	if field == nil || s.responseModel == nil {
		return true
	}

	return fieldIncludedInResponse(s.responseModel, field, s.responseByMask)
}

func (s *QueryRunner) prefetchRelations(modelList []models.Model, deepLevel int) (err *cd.Error) {
	if len(modelList) <= 1 || deepLevel > maxDeepLevel {
		return
	}

	groupMap := map[string]*relationPrefetchGroup{}
	groupIndex := []string{}
	for _, modelVal := range modelList {
		if modelVal == nil {
			continue
		}

		pkField := modelVal.GetPrimaryField()
		if pkField == nil || !models.IsAssignedField(pkField) {
			continue
		}

		leftID, leftErr := s.modelCodec.ExtractBasicFieldValue(pkField, pkField.GetValue().Get())
		if leftErr != nil {
			err = leftErr
			return
		}

		for _, field := range modelVal.GetFields() {
			if models.IsBasicField(field) || !s.shouldLoadRelationField(field) {
				continue
			}

			groupKey := fmt.Sprintf("%s:%s", modelVal.GetPkgKey(), field.GetName())
			group := groupMap[groupKey]
			if group == nil {
				group = &relationPrefetchGroup{model: modelVal, field: field}
				groupMap[groupKey] = group
				groupIndex = append(groupIndex, groupKey)
			}

			group.leftIDs = append(group.leftIDs, leftID)
		}
	}

	for _, groupKey := range groupIndex {
		group := groupMap[groupKey]
		if group == nil || len(group.leftIDs) == 0 {
			continue
		}
		if err = s.batchQueryRelationKeys(group.model, group.field, group.leftIDs); err != nil {
			return
		}
	}

	for _, groupKey := range groupIndex {
		group := groupMap[groupKey]
		if group == nil || len(group.leftIDs) == 0 {
			continue
		}

		relationType := group.field.GetType()
		if models.IsSliceField(group.field) {
			relationType = relationType.Elem()
		}

		uniqueIDs := []any{}
		seenIDs := map[string]struct{}{}
		for _, leftID := range group.leftIDs {
			rightIDs, ok := s.getCachedRelationEdge(group.model.GetPkgKey(), group.field.GetName(), leftID)
			if !ok {
				continue
			}

			for _, rightID := range rightIDs {
				rightKey := fmt.Sprintf("%v", rightID)
				if _, exists := seenIDs[rightKey]; exists {
					continue
				}
				seenIDs[rightKey] = struct{}{}
				uniqueIDs = append(uniqueIDs, rightID)
			}
		}

		if len(uniqueIDs) == 0 {
			continue
		}
		if err = s.batchQueryRelationModels(relationType, uniqueIDs, deepLevel); err != nil {
			return
		}
	}

	return
}

func (s *QueryRunner) batchQueryRelationKeys(vModel models.Model, vField models.Field, leftIDs []any) (err *cd.Error) {
	if len(leftIDs) == 0 {
		return
	}

	pkField := vModel.GetPrimaryField()
	if pkField == nil {
		err = cd.NewError(cd.Unexpected, "relation owner model missing primary field")
		return
	}

	orderedIDs := []any{}
	seenIDs := map[string]struct{}{}
	for _, leftID := range leftIDs {
		normalizedID, normalizedErr := s.modelCodec.ExtractBasicFieldValue(pkField, leftID)
		if normalizedErr != nil {
			err = normalizedErr
			return
		}

		edgeKey := relationEdgeKey(vModel.GetPkgKey(), vField.GetName(), normalizedID)
		if _, exists := s.relationEdges[edgeKey]; exists {
			continue
		}
		if _, exists := seenIDs[edgeKey]; exists {
			continue
		}

		seenIDs[edgeKey] = struct{}{}
		orderedIDs = append(orderedIDs, normalizedID)
		s.cacheRelationEdge(vModel.GetPkgKey(), vField.GetName(), normalizedID, nil)
	}
	if len(orderedIDs) == 0 {
		return
	}

	relationResult, relationErr := s.sqlBuilder.BuildBatchQueryRelation(vModel, vField, orderedIDs)
	if relationErr != nil {
		err = relationErr
		return
	}

	relationType := vField.GetType()
	if models.IsSliceField(vField) {
		relationType = relationType.Elem()
	}

	relationModel, modelErr := s.modelProvider.GetTypeModel(relationType)
	if modelErr != nil {
		err = modelErr
		return
	}
	relationPK := relationModel.GetPrimaryField()
	if relationPK == nil {
		err = cd.NewError(cd.Unexpected, "relation model missing primary field")
		return
	}

	_, err = s.executor.Query(relationResult.SQL(), false, relationResult.Args()...)
	if err != nil {
		return
	}
	defer s.executor.Finish()

	for s.executor.Next() {
		var leftID any
		var rightID any
		if err = s.executor.GetField(&leftID, &rightID); err != nil {
			return
		}

		normalizedLeftID, leftErr := s.modelCodec.ExtractBasicFieldValue(pkField, leftID)
		if leftErr != nil {
			err = leftErr
			return
		}
		normalizedRightID, rightErr := s.modelCodec.ExtractBasicFieldValue(relationPK, rightID)
		if rightErr != nil {
			err = rightErr
			return
		}

		cachedIDs, _ := s.getCachedRelationEdge(vModel.GetPkgKey(), vField.GetName(), normalizedLeftID)
		cachedIDs = append(cachedIDs, normalizedRightID)
		s.cacheRelationEdge(vModel.GetPkgKey(), vField.GetName(), normalizedLeftID, cachedIDs)
	}

	return
}

func (s *QueryRunner) getRelationIDs(vModel models.Model, vField models.Field) (ret resultItems, cached bool, err *cd.Error) {
	pkField := vModel.GetPrimaryField()
	if pkField != nil && models.IsAssignedField(pkField) {
		leftID, leftErr := s.modelCodec.ExtractBasicFieldValue(pkField, pkField.GetValue().Get())
		if leftErr != nil {
			err = leftErr
			return
		}

		if edgeIDs, ok := s.getCachedRelationEdge(vModel.GetPkgKey(), vField.GetName(), leftID); ok {
			ret = append(resultItems{}, edgeIDs...)
			cached = true
			return
		}
	}

	ret, err = s.innerQueryRelationKeys(vModel, vField)
	return
}

func (s *QueryRunner) batchQueryRelationModels(relationType models.Type, ids []any, deepLevel int) (err *cd.Error) {
	if len(ids) == 0 {
		return
	}

	rModel, rErr := s.modelProvider.GetTypeModel(relationType)
	if rErr != nil {
		err = rErr
		return
	}

	pkField := rModel.GetPrimaryField()
	if pkField == nil {
		err = cd.NewError(cd.Unexpected, "relation model missing primary field")
		return
	}

	vFilter, vErr := s.modelProvider.GetModelFilter(rModel)
	if vErr != nil {
		err = vErr
		return
	}
	if err = vFilter.In(pkField.GetName(), buildTypedValueSlice(ids)); err != nil {
		return
	}

	queryMask, maskErr := buildFullQueryMaskModel(rModel)
	if maskErr != nil {
		err = maskErr
		return
	}

	rQueryRunner := NewQueryRunner(s.context, queryMask, queryMask, true, s.executor, s.modelProvider, s.modelCodec, true, deepLevel+1)
	rQueryRunner.relationCache = s.relationCache
	rQueryRunner.relationMisses = s.relationMisses
	rQueryRunner.relationEdges = s.relationEdges
	queryVal, queryErr := rQueryRunner.Query(vFilter)
	if queryErr != nil {
		err = queryErr
		return
	}

	foundIDs := map[string]struct{}{}
	for _, modelVal := range queryVal {
		s.cacheRelationModel(modelVal)
		pkValue := modelVal.GetPrimaryField()
		if pkValue == nil || !models.IsAssignedField(pkValue) {
			continue
		}
		normalizedID, normalizedErr := s.modelCodec.ExtractBasicFieldValue(pkField, pkValue.GetValue().Get())
		if normalizedErr != nil {
			err = normalizedErr
			return
		}
		foundIDs[fmt.Sprintf("%v", normalizedID)] = struct{}{}
	}

	for _, id := range ids {
		normalizedID, normalizedErr := s.modelCodec.ExtractBasicFieldValue(pkField, id)
		if normalizedErr != nil {
			err = normalizedErr
			return
		}
		if _, ok := foundIDs[fmt.Sprintf("%v", normalizedID)]; ok {
			continue
		}
		s.cacheRelationMiss(rModel.GetPkgKey(), normalizedID)
	}

	return
}

func (s *QueryRunner) prepareRelationIDs(relationModel models.Model, ids []any) (ordered []any, pending []any, err *cd.Error) {
	pkField := relationModel.GetPrimaryField()
	if pkField == nil {
		err = cd.NewError(cd.Unexpected, "relation model missing primary field")
		return
	}

	pendingIndex := map[string]struct{}{}
	for _, id := range ids {
		rVal, rErr := s.modelCodec.ExtractBasicFieldValue(pkField, id)
		if rErr != nil {
			err = rErr
			return
		}

		ordered = append(ordered, rVal)
		cacheKey := relationCacheKey(relationModel.GetPkgKey(), rVal)
		if s.relationCache != nil && s.relationCache[cacheKey] != nil {
			continue
		}
		if s.isRelationMiss(relationModel.GetPkgKey(), rVal) {
			continue
		}
		if _, exists := pendingIndex[cacheKey]; exists {
			continue
		}

		pending = append(pending, rVal)
		pendingIndex[cacheKey] = struct{}{}
	}

	return
}

func (s *QueryRunner) queryRelationModel(relationType models.Type, id any, deepLevel int) (ret models.Model, err *cd.Error) {
	rModel, rErr := s.modelProvider.GetTypeModel(relationType)
	if rErr != nil {
		err = rErr
		return
	}

	rVal, rErr := s.modelCodec.ExtractBasicFieldValue(rModel.GetPrimaryField(), id)
	if rErr != nil {
		err = rErr
		return
	}
	if cachedModel := s.getCachedRelationModel(rModel.GetPkgKey(), rVal); cachedModel != nil {
		ret = cachedModel
		return
	}
	if s.isRelationMiss(rModel.GetPkgKey(), rVal) {
		return
	}

	rModel.SetPrimaryFieldValue(rVal)
	vFilter, vErr := getModelFilter(rModel, s.modelProvider, s.modelCodec)
	if vErr != nil {
		err = vErr
		return
	}

	queryMask, maskErr := buildFullQueryMaskModel(rModel)
	if maskErr != nil {
		err = maskErr
		return
	}

	rQueryRunner := NewQueryRunner(s.context, queryMask, queryMask, false, s.executor, s.modelProvider, s.modelCodec, false, deepLevel+1)
	rQueryRunner.relationCache = s.relationCache
	rQueryRunner.relationMisses = s.relationMisses
	rQueryRunner.relationEdges = s.relationEdges
	queryVal, queryErr := rQueryRunner.Query(vFilter)
	if queryErr != nil {
		err = queryErr
		return
	}
	if len(queryVal) > 1 {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("match more than one model, model:%s, id:%v", rModel.GetPkgKey(), id))
		return
	}
	if len(queryVal) == 0 {
		s.cacheRelationMiss(rModel.GetPkgKey(), rVal)
		return
	}

	s.cacheRelationModel(queryVal[0])
	ret = queryVal[0]
	return
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

func (s *QueryRunner) innerAssignBasic(vModel models.Model, queryVal resultItems) (ret models.Model, err *cd.Error) {
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

	ret = qModel
	return
}

func (s *QueryRunner) innerAssignRelations(qModel models.Model, deepLevel int) (err *cd.Error) {
	for _, field := range qModel.GetFields() {
		if models.IsBasicField(field) || !s.shouldLoadRelationField(field) {
			continue
		}
		err = s.assignModelField(qModel, field, deepLevel)
		if err != nil {
			slog.Error("QueryRunner failed", "error", err.Error())
			return
		}
	}

	return
}

func (s *QueryRunner) innerAssign(vModel models.Model, queryVal resultItems, deepLevel int) (ret models.Model, err *cd.Error) {
	qModel, assignErr := s.innerAssignBasic(vModel, queryVal)
	if assignErr != nil {
		err = assignErr
		return
	}
	if err = s.innerAssignRelations(qModel, deepLevel); err != nil {
		return
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
	valueList, _, valueErr := s.getRelationIDs(vModel, vField)
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
	valueList, _, valueErr := s.getRelationIDs(vModel, vField)
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
	rModel, rErr := s.queryRelationModel(vField.GetType(), id, deepLevel)
	if rErr != nil {
		err = rErr
		slog.Error("QueryRunner assignBasicField failed", "fieldId", id, "error", err.Error())
		return
	}

	if rModel != nil {
		vField.SetValue(rModel.Interface(vField.GetType().Elem().IsPtrType()))
		return
	}

	if deepLevel < maxDeepLevel {
		// 到这里说明未查询到关联目标，当前行为是记录告警并保留字段为空，
		// 由调用方或外部治理流程处理关系表与目标表之间的数据不一致。
		slog.Warn("query relation failed, miss relation data", "model", vField.GetType().GetPkgKey(), "id", id)
	}
	return
}

func (s *QueryRunner) innerQueryRelationSliceModel(ids []any, vField models.Field, deepLevel int) (err *cd.Error) {
	// 这里主动重置，避免VFiled的旧数据干扰
	vField.Reset()
	svModel, svErr := s.modelProvider.GetTypeModel(vField.GetType().Elem())
	if svErr != nil {
		err = svErr
		slog.Error("QueryRunner failed", "error", err.Error())
		return
	}

	pkField := svModel.GetPrimaryField()
	if pkField == nil {
		err = cd.NewError(cd.Unexpected, "relation model missing primary field")
		slog.Error("QueryRunner failed", "error", err.Error())
		return
	}

	idOrder, missingIDs, prepareErr := s.prepareRelationIDs(svModel, ids)
	if prepareErr != nil {
		err = prepareErr
		slog.Error("QueryRunner failed", "error", err.Error())
		return
	}

	if len(missingIDs) > 0 {
		if err = s.batchQueryRelationModels(vField.GetType().Elem(), missingIDs, deepLevel); err != nil {
			slog.Error("QueryRunner failed", "error", err.Error())
			return
		}
	}

	for _, id := range idOrder {
		cachedModel := s.getCachedRelationModel(svModel.GetPkgKey(), id)
		if cachedModel == nil && !s.isRelationMiss(svModel.GetPkgKey(), id) {
			var queryErr *cd.Error
			cachedModel, queryErr = s.queryRelationModel(vField.GetType().Elem(), id, deepLevel)
			if queryErr != nil {
				err = queryErr
				slog.Error("QueryRunner failed", "error", err.Error())
				return
			}
		}
		if cachedModel == nil {
			if deepLevel < maxDeepLevel {
				slog.Warn("query relation failed, miss relation data", "model", svModel.GetPkgKey(), "id", id)
			}
			continue
		}

		vField.AppendSliceValue(cachedModel.Interface(vField.GetType().Elem().IsPtrType()))
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
		modelVal, modelErr := s.innerAssignBasic(s.vModel, queryValueList[idx])
		if modelErr != nil {
			err = modelErr
			slog.Error("QueryRunner failed", "error", err.Error())
			return
		}
		sliceValue = append(sliceValue, modelVal)
	}

	if err = s.prefetchRelations(sliceValue, s.deepLevel); err != nil {
		slog.Error("QueryRunner prefetchRelations failed", "error", err.Error())
		return
	}

	for idx := range sliceValue {
		if err = s.innerAssignRelations(sliceValue[idx], s.deepLevel); err != nil {
			slog.Error("QueryRunner failed", "error", err.Error())
			return
		}
		sliceValue[idx] = applyQueryResponseModel(sliceValue[idx], s.responseModel, s.responseByMask)
	}

	ret = sliceValue
	return
}

func (s *impl) Query(vModel models.Model) (ret models.Model, err *cd.Error) {
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		if ormMetricCollector != nil {
			ormMetricCollector.RecordOperation(string(metrics.OperationQuery), vModel, duration, cd.ToStdError(err))
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
