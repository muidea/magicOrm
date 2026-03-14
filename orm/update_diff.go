package orm

import (
	"fmt"
	"log/slog"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
)

// normalizeID 将主键值统一为字符串以便差集比较
func normalizeID(id any) string {
	return fmt.Sprintf("%v", id)
}

// diffRelationIDs 求关系 right ID 的差集：toDelete = existing - new，toInsert = new - existing
func diffRelationIDs(existing, new []any) (toDelete, toInsert []any) {
	existingSet := make(map[string]any, len(existing))
	for _, id := range existing {
		existingSet[normalizeID(id)] = id
	}
	newSet := make(map[string]any, len(new))
	for _, id := range new {
		newSet[normalizeID(id)] = id
	}
	for key, id := range existingSet {
		if _, found := newSet[key]; !found {
			toDelete = append(toDelete, id)
		}
	}
	for key, id := range newSet {
		if _, found := existingSet[key]; !found {
			toInsert = append(toInsert, id)
		}
	}
	return
}

// prepareNewRightIDs 从当前 Model 的关系字段值中提取要写入的 right 主键列表（引用关系）
func (s *UpdateRunner) prepareNewRightIDs(vModel models.Model, vField models.Field) (ret []any, err *cd.Error) {
	elemType := vField.GetType().Elem()

	if models.IsSliceField(vField) {
		fSliceValue := vField.GetSliceValue()
		if len(fSliceValue) == 0 {
			ret = []any{}
			return
		}
		newRightIDs := make([]any, 0, len(fSliceValue))
		for _, fVal := range fSliceValue {
			rModel, rErr := s.modelProvider.GetTypeModel(elemType)
			if rErr != nil {
				err = rErr
				slog.Error("prepareNewRightIDs GetTypeModel failed", "field", vField.GetName(), "error", err.Error())
				return
			}
			rModel, rErr = s.modelProvider.SetModelValue(rModel, fVal)
			if rErr != nil {
				err = rErr
				slog.Error("prepareNewRightIDs SetModelValue failed", "field", vField.GetName(), "error", err.Error())
				return
			}
			pkField := rModel.GetPrimaryField()
			pkValue := pkField.GetValue()
			if pkValue.IsZero() {
				err = cd.NewError(cd.IllegalParam, fmt.Sprintf("reference relation field %s has entity without primary key", vField.GetName()))
				slog.Error("prepareNewRightIDs: reference entity missing primary key", "field", vField.GetName())
				return
			}
			encodedPK, encErr := s.modelCodec.PackedBasicFieldValue(pkField, pkValue)
			if encErr != nil {
				err = encErr
				slog.Error("prepareNewRightIDs PackedBasicFieldValue failed", "field", vField.GetName(), "error", err.Error())
				return
			}
			newRightIDs = append(newRightIDs, encodedPK)
		}
		ret = newRightIDs
		return
	}

	// 单值关系：nil/零值表示清空引用，返回空列表
	if vField.GetValue().IsZero() {
		ret = []any{}
		return
	}
	rModel, rErr := s.modelProvider.GetTypeModel(elemType)
	if rErr != nil {
		err = rErr
		slog.Error("prepareNewRightIDs GetTypeModel failed", "field", vField.GetName(), "error", err.Error())
		return
	}
	rModel, rErr = s.modelProvider.SetModelValue(rModel, vField.GetValue())
	if rErr != nil {
		err = rErr
		slog.Error("prepareNewRightIDs SetModelValue failed", "field", vField.GetName(), "error", err.Error())
		return
	}
	pkField := rModel.GetPrimaryField()
	pkValue := pkField.GetValue()
	if pkValue.IsZero() {
		err = cd.NewError(cd.IllegalParam, fmt.Sprintf("reference relation field %s has entity without primary key", vField.GetName()))
		slog.Error("prepareNewRightIDs: reference entity missing primary key", "field", vField.GetName())
		return
	}
	encodedPK, encErr := s.modelCodec.PackedBasicFieldValue(pkField, pkValue)
	if encErr != nil {
		err = encErr
		slog.Error("prepareNewRightIDs PackedBasicFieldValue failed", "field", vField.GetName(), "error", err.Error())
		return
	}
	ret = []any{encodedPK}
	return
}

// queryExistingRightIDs 查询关系表中当前该 host 字段已有的 right ID 列表
func (s *UpdateRunner) queryExistingRightIDs(vModel models.Model, vField models.Field) (ret []any, err *cd.Error) {
	existingIDs, queryErr := s.innerQueryRelationKeys(vModel, vField)
	if queryErr != nil {
		err = queryErr
		slog.Error("queryExistingRightIDs innerQueryRelationKeys failed", "field", vField.GetName(), "error", err.Error())
		return
	}
	ret = existingIDs
	return
}

// deleteRelationLinks 仅删除关系表中指定的 (left, right) 行，不删除关联实体
func (s *UpdateRunner) deleteRelationLinks(vModel models.Model, vField models.Field, toDelete []any) (err *cd.Error) {
	if len(toDelete) == 0 {
		return
	}
	result, buildErr := s.sqlBuilder.BuildDeleteRelationByRights(vModel, vField, toDelete)
	if buildErr != nil {
		err = buildErr
		slog.Error("deleteRelationLinks BuildDeleteRelationByRights failed", "field", vField.GetName(), "error", err.Error())
		return
	}
	if result == nil {
		return
	}
	_, err = s.executor.Execute(result.SQL(), result.Args()...)
	if err != nil {
		slog.Error("deleteRelationLinks Execute failed", "field", vField.GetName(), "error", err.Error())
	}
	return
}

// insertRelationLinks 仅向关系表插入指定的 (left, right) 行
func (s *UpdateRunner) insertRelationLinks(vModel models.Model, vField models.Field, toInsert []any) (err *cd.Error) {
	if len(toInsert) == 0 {
		return
	}
	elemType := vField.GetType().Elem()
	for _, rightID := range toInsert {
		rModel, rErr := s.modelProvider.GetTypeModel(elemType)
		if rErr != nil {
			err = rErr
			slog.Error("insertRelationLinks GetTypeModel failed", "field", vField.GetName(), "error", err.Error())
			return
		}
		rVal, rErr := s.modelCodec.ExtractBasicFieldValue(rModel.GetPrimaryField(), rightID)
		if rErr != nil {
			err = rErr
			slog.Error("insertRelationLinks ExtractBasicFieldValue failed", "field", vField.GetName(), "error", err.Error())
			return
		}
		if setErr := rModel.SetPrimaryFieldValue(rVal); setErr != nil {
			err = setErr
			slog.Error("insertRelationLinks SetPrimaryFieldValue failed", "field", vField.GetName(), "error", err.Error())
			return
		}
		relationResult, relationErr := s.sqlBuilder.BuildInsertRelation(vModel, vField, rModel)
		if relationErr != nil {
			err = relationErr
			slog.Error("insertRelationLinks BuildInsertRelation failed", "field", vField.GetName(), "error", err.Error())
			return
		}
		var idVal any
		err = s.executor.ExecuteInsert(relationResult.SQL(), &idVal, relationResult.Args()...)
		if err != nil {
			slog.Error("insertRelationLinks ExecuteInsert failed", "field", vField.GetName(), "error", err.Error())
			return
		}
	}
	return
}

// updateReferenceRelation 引用关系：只刷新关系（增删链接），不处理实体
func (s *UpdateRunner) updateReferenceRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
	newRightIDs, prepErr := s.prepareNewRightIDs(vModel, vField)
	if prepErr != nil {
		err = prepErr
		slog.Error("UpdateRunner updateReferenceRelation prepareNewRightIDs failed", "field", vField.GetName(), "error", err.Error())
		return
	}
	existingRightIDs, queryErr := s.queryExistingRightIDs(vModel, vField)
	if queryErr != nil {
		err = queryErr
		slog.Error("UpdateRunner updateReferenceRelation queryExistingRightIDs failed", "field", vField.GetName(), "error", err.Error())
		return
	}
	toDelete, toInsert := diffRelationIDs(existingRightIDs, newRightIDs)
	err = s.deleteRelationLinks(vModel, vField, toDelete)
	if err != nil {
		slog.Error("UpdateRunner updateReferenceRelation deleteRelationLinks failed", "field", vField.GetName(), "error", err.Error())
		return
	}
	err = s.insertRelationLinks(vModel, vField, toInsert)
	if err != nil {
		slog.Error("UpdateRunner updateReferenceRelation insertRelationLinks failed", "field", vField.GetName(), "error", err.Error())
	}
	return
}

// updateContainRelation 包含关系：同步处理关系和实体（先 deleteRelation 再 insertRelation）
func (s *UpdateRunner) updateContainRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
	existingField, queryErr := s.queryExistingContainRelationField(vModel, vField)
	if queryErr != nil {
		err = queryErr
		slog.Error("UpdateRunner updateContainRelation queryExistingContainRelationField failed", "field", vField.GetName(), "error", err.Error())
		return
	}

	unchanged, compareErr := s.compareRelationFieldValue(existingField, vField, 0)
	if compareErr != nil {
		err = compareErr
		slog.Error("UpdateRunner updateContainRelation isContainRelationUnchanged failed", "field", vField.GetName(), "error", err.Error())
		return
	}
	if unchanged {
		return
	}

	handled, preciseErr := s.updateContainRelationPrecisely(vModel, existingField, vField)
	if preciseErr != nil {
		err = preciseErr
		slog.Error("UpdateRunner updateContainRelation updateContainRelationPrecisely failed", "field", vField.GetName(), "error", err.Error())
		return
	}
	if handled {
		return
	}

	newVal := vField.GetValue().Get()
	err = s.deleteRelation(vModel, vField, 0)
	if err != nil {
		slog.Error("UpdateRunner updateContainRelation deleteRelation failed", "field", vField.GetName(), "error", err.Error())
		return
	}
	if !models.IsValidField(vField) {
		return
	}
	vField.SetValue(newVal)
	err = s.insertRelation(vModel, vField)
	if err != nil {
		slog.Error("UpdateRunner updateContainRelation insertRelation failed", "field", vField.GetName(), "error", err.Error())
	}
	return
}

func (s *UpdateRunner) isContainRelationUnchanged(vModel models.Model, vField models.Field) (ret bool, err *cd.Error) {
	existingField, queryErr := s.queryExistingContainRelationField(vModel, vField)
	if queryErr != nil {
		err = queryErr
		return
	}

	ret, err = s.compareRelationFieldValue(existingField, vField, 0)
	return
}

func (s *UpdateRunner) updateContainRelationPrecisely(vModel models.Model, existingField, newField models.Field) (ret bool, err *cd.Error) {
	if models.IsSliceField(newField) {
		ret, err = s.updateContainSliceRelationPrecisely(vModel, existingField, newField)
		return
	}

	ret, err = s.updateContainSingleRelationPrecisely(vModel, existingField, newField)
	return
}

func (s *UpdateRunner) queryExistingContainRelationField(vModel models.Model, vField models.Field) (ret models.Field, err *cd.Error) {
	queryModel := vModel.Copy(models.OriginView)
	pkField := vModel.GetPrimaryField()
	if pkField == nil || pkField.GetValue() == nil || !pkField.GetValue().IsValid() {
		err = cd.NewError(cd.IllegalParam, fmt.Sprintf("contain relation field %s requires primary key", vField.GetName()))
		return
	}
	if setErr := queryModel.SetPrimaryFieldValue(pkField.GetValue().Get()); setErr != nil {
		err = setErr
		return
	}

	queryField := queryModel.GetField(vField.GetName())
	if queryField == nil {
		err = cd.NewError(cd.IllegalParam, fmt.Sprintf("contain relation field %s not found", vField.GetName()))
		return
	}

	if queryErr := s.queryRelation(queryModel, queryField, 0); queryErr != nil {
		err = queryErr
		return
	}

	ret = queryField
	return
}

func (s *UpdateRunner) compareRelationFieldValue(existingField, newField models.Field, depth int) (ret bool, err *cd.Error) {
	if depth > maxDeepLevel {
		return
	}

	if models.IsSliceField(newField) {
		ret, err = s.compareRelationSliceFieldValue(existingField, newField, depth)
		return
	}

	ret, err = s.compareRelationSingleFieldValue(existingField, newField, depth)
	return
}

func (s *UpdateRunner) compareRelationSingleFieldValue(existingField, newField models.Field, depth int) (ret bool, err *cd.Error) {
	if !models.IsValidField(existingField) || !models.IsValidField(newField) {
		ret = !models.IsValidField(existingField) && !models.IsValidField(newField)
		return
	}

	existingRaw := existingField.GetValue().Get()
	newRaw := newField.GetValue().Get()
	if existingRaw == nil || newRaw == nil {
		ret = utils.IsSameValue(existingRaw, newRaw)
		return
	}

	existingModel, modelErr := s.relationModelFromRaw(existingField.GetType(), existingRaw)
	if modelErr != nil {
		err = modelErr
		return
	}
	newModel, modelErr := s.relationModelFromRaw(newField.GetType(), newRaw)
	if modelErr != nil {
		err = modelErr
		return
	}

	ret, err = s.compareRelationModelValue(existingModel, newModel, depth+1)
	return
}

func (s *UpdateRunner) compareRelationSliceFieldValue(existingField, newField models.Field, depth int) (ret bool, err *cd.Error) {
	existingValues := existingField.GetSliceValue()
	newValues := newField.GetSliceValue()
	if len(existingValues) != len(newValues) {
		return
	}

	existingModels, modelErr := s.relationModelsFromSliceValue(existingField.GetType().Elem(), existingValues)
	if modelErr != nil {
		err = modelErr
		return
	}
	newModels, modelErr := s.relationModelsFromSliceValue(newField.GetType().Elem(), newValues)
	if modelErr != nil {
		err = modelErr
		return
	}

	if s.canCompareRelationModelsByPrimary(existingModels) && s.canCompareRelationModelsByPrimary(newModels) {
		ret, err = s.compareRelationModelSliceByPrimary(existingModels, newModels, depth+1)
		return
	}

	for idx := range existingModels {
		same, compareErr := s.compareRelationModelValue(existingModels[idx], newModels[idx], depth+1)
		if compareErr != nil {
			err = compareErr
			return
		}
		if !same {
			return
		}
	}

	ret = true
	return
}

func (s *UpdateRunner) compareRelationModelSliceByPrimary(existingModels, newModels []models.Model, depth int) (ret bool, err *cd.Error) {
	newModelMap := make(map[string]models.Model, len(newModels))
	for _, modelVal := range newModels {
		pkVal, ok := getModelPrimaryKey(modelVal)
		if !ok {
			return
		}
		if _, found := newModelMap[pkVal]; found {
			return
		}
		newModelMap[pkVal] = modelVal
	}

	if len(newModelMap) != len(existingModels) {
		return
	}

	for _, modelVal := range existingModels {
		pkVal, ok := getModelPrimaryKey(modelVal)
		if !ok {
			return
		}
		targetModel, found := newModelMap[pkVal]
		if !found {
			return
		}

		same, compareErr := s.compareRelationModelValue(modelVal, targetModel, depth+1)
		if compareErr != nil {
			err = compareErr
			return
		}
		if !same {
			return
		}
	}

	ret = true
	return
}

func (s *UpdateRunner) compareRelationModelValue(existingModel, newModel models.Model, depth int) (ret bool, err *cd.Error) {
	if depth > maxDeepLevel {
		return
	}
	if existingModel == nil || newModel == nil {
		ret = existingModel == nil && newModel == nil
		return
	}
	if existingModel.GetPkgKey() != newModel.GetPkgKey() {
		return
	}

	for _, existingField := range existingModel.GetFields() {
		newField := newModel.GetField(existingField.GetName())
		if newField == nil {
			return
		}

		if models.IsBasicField(existingField) {
			if !utils.IsSameValue(existingField.GetValue().Get(), newField.GetValue().Get()) {
				return
			}
			continue
		}

		same, compareErr := s.compareRelationFieldValue(existingField, newField, depth+1)
		if compareErr != nil {
			err = compareErr
			return
		}
		if !same {
			return
		}
	}

	ret = true
	return
}

type containModelEntry struct {
	model      models.Model
	rawValue   models.Value
	hasPrimary bool
	primaryID  string
	packedID   any
}

func (s *UpdateRunner) updateContainSingleRelationPrecisely(vModel models.Model, existingField, newField models.Field) (ret bool, err *cd.Error) {
	if !models.IsValidField(existingField) || !models.IsValidField(newField) {
		return
	}

	existingRaw := existingField.GetValue().Get()
	newRaw := newField.GetValue().Get()
	if existingRaw == nil || newRaw == nil {
		return
	}

	existingModel, modelErr := s.relationModelFromRaw(existingField.GetType(), existingRaw)
	if modelErr != nil {
		err = modelErr
		return
	}
	newModel, modelErr := s.relationModelFromRaw(newField.GetType(), newRaw)
	if modelErr != nil {
		err = modelErr
		return
	}

	existingPK, existingOK := getModelPrimaryKey(existingModel)
	newPK, newOK := getModelPrimaryKey(newModel)
	if !existingOK || !newOK || existingPK != newPK {
		return
	}

	updatedModel, updateErr := NewUpdateRunner(s.context, newModel, s.executor, s.modelProvider, s.modelCodec).Update()
	if updateErr != nil {
		err = updateErr
		return
	}
	if setErr := newField.SetValue(updatedModel.Interface(models.IsPtrField(newField))); setErr != nil {
		err = setErr
		return
	}

	ret = true
	return
}

func (s *UpdateRunner) updateContainSliceRelationPrecisely(vModel models.Model, existingField, newField models.Field) (ret bool, err *cd.Error) {
	existingEntries, buildErr := s.buildContainEntries(existingField.GetType().Elem(), existingField.GetSliceValue())
	if buildErr != nil {
		err = buildErr
		return
	}
	newEntries, buildErr := s.buildContainEntries(newField.GetType().Elem(), newField.GetSliceValue())
	if buildErr != nil {
		err = buildErr
		return
	}

	existingMap, buildOK := buildContainPrimaryEntryMap(existingEntries)
	if !buildOK {
		return
	}
	newMap, pendingInsertEntries, buildOK := buildContainPrimaryEntryMapForUpdate(newEntries)
	if !buildOK {
		return
	}

	toDeleteIDs := []any{}
	for primaryID, existingEntry := range existingMap {
		newEntry, found := newMap[primaryID]
		if !found {
			if deleteErr := s.deleteContainedModel(existingEntry.model); deleteErr != nil {
				err = deleteErr
				return
			}
			toDeleteIDs = append(toDeleteIDs, existingEntry.packedID)
			continue
		}

		same, compareErr := s.compareRelationModelValue(existingEntry.model, newEntry.model, 1)
		if compareErr != nil {
			err = compareErr
			return
		}
		if same {
			continue
		}

		updatedModel, updateErr := NewUpdateRunner(s.context, newEntry.model, s.executor, s.modelProvider, s.modelCodec).Update()
		if updateErr != nil {
			err = updateErr
			return
		}
		if setErr := newEntry.rawValue.Set(updatedModel.Interface(newField.GetType().Elem().IsPtrType())); setErr != nil {
			err = setErr
			return
		}
	}

	if len(toDeleteIDs) > 0 {
		if deleteErr := s.deleteRelationLinks(vModel, newField, toDeleteIDs); deleteErr != nil {
			err = deleteErr
			return
		}
	}

	for _, newEntry := range pendingInsertEntries {
		insertedModel, insertErr := s.insertContainedModel(vModel, newField, newEntry.model)
		if insertErr != nil {
			err = insertErr
			return
		}
		if setErr := newEntry.rawValue.Set(insertedModel.Interface(newField.GetType().Elem().IsPtrType())); setErr != nil {
			err = setErr
			return
		}
	}

	for _, newEntry := range newEntries {
		if newEntry.hasPrimary {
			if _, found := existingMap[newEntry.primaryID]; found {
				continue
			}
			insertedModel, insertErr := s.insertContainedModel(vModel, newField, newEntry.model)
			if insertErr != nil {
				err = insertErr
				return
			}
			if setErr := newEntry.rawValue.Set(insertedModel.Interface(newField.GetType().Elem().IsPtrType())); setErr != nil {
				err = setErr
				return
			}
		}
	}

	ret = true
	return
}

func (s *UpdateRunner) deleteContainedModel(vModel models.Model) (err *cd.Error) {
	err = NewDeleteRunner(s.context, vModel, s.executor, s.modelProvider, s.modelCodec, 1).Delete()
	return
}

func (s *UpdateRunner) insertContainedModel(parentModel models.Model, parentField models.Field, childModel models.Model) (ret models.Model, err *cd.Error) {
	insertedModel, insertErr := NewInsertRunner(s.context, childModel, s.executor, s.modelProvider, s.modelCodec).Insert()
	if insertErr != nil {
		err = insertErr
		return
	}

	relationResult, relationErr := s.sqlBuilder.BuildInsertRelation(parentModel, parentField, insertedModel)
	if relationErr != nil {
		err = relationErr
		return
	}

	var idVal any
	err = s.executor.ExecuteInsert(relationResult.SQL(), &idVal, relationResult.Args()...)
	if err != nil {
		return
	}

	ret = insertedModel
	return
}

func (s *UpdateRunner) relationModelFromRaw(tType models.Type, raw any) (ret models.Model, err *cd.Error) {
	modelVal, modelErr := s.modelProvider.GetTypeModel(tType)
	if modelErr != nil {
		err = modelErr
		return
	}

	entityValue, valueErr := s.modelProvider.GetEntityValue(raw)
	if valueErr != nil {
		err = valueErr
		return
	}

	ret, err = s.modelProvider.SetModelValue(modelVal, entityValue)
	return
}

func (s *UpdateRunner) relationModelsFromSliceValue(elemType models.Type, rawValues []models.Value) (ret []models.Model, err *cd.Error) {
	ret = make([]models.Model, 0, len(rawValues))
	for _, rawValue := range rawValues {
		if rawValue == nil || !rawValue.IsValid() {
			continue
		}

		modelVal, modelErr := s.relationModelFromRaw(elemType, rawValue.Get())
		if modelErr != nil {
			err = modelErr
			return
		}
		ret = append(ret, modelVal)
	}
	return
}

func (s *UpdateRunner) buildContainEntries(elemType models.Type, rawValues []models.Value) (ret []containModelEntry, err *cd.Error) {
	ret = make([]containModelEntry, 0, len(rawValues))
	for _, rawValue := range rawValues {
		if rawValue == nil || !rawValue.IsValid() {
			continue
		}

		modelVal, modelErr := s.relationModelFromRaw(elemType, rawValue.Get())
		if modelErr != nil {
			err = modelErr
			return
		}

		entry := containModelEntry{
			model:    modelVal,
			rawValue: rawValue,
		}
		entry.primaryID, entry.hasPrimary = getModelPrimaryKey(modelVal)
		if entry.hasPrimary {
			entry.packedID, err = s.packModelPrimaryID(modelVal)
			if err != nil {
				return
			}
		}
		ret = append(ret, entry)
	}
	return
}

func (s *UpdateRunner) canCompareRelationModelsByPrimary(modelList []models.Model) bool {
	if len(modelList) == 0 {
		return true
	}

	seenPrimary := make(map[string]struct{}, len(modelList))
	for _, modelVal := range modelList {
		pkVal, ok := getModelPrimaryKey(modelVal)
		if !ok {
			return false
		}
		if _, found := seenPrimary[pkVal]; found {
			return false
		}
		seenPrimary[pkVal] = struct{}{}
	}
	return true
}

func getModelPrimaryKey(vModel models.Model) (ret string, ok bool) {
	if vModel == nil {
		return
	}

	pkField := vModel.GetPrimaryField()
	if pkField == nil || pkField.GetValue() == nil || !pkField.GetValue().IsValid() || pkField.GetValue().IsZero() {
		return
	}

	ret = normalizeID(pkField.GetValue().Get())
	ok = true
	return
}

func (s *UpdateRunner) packModelPrimaryID(vModel models.Model) (ret any, err *cd.Error) {
	pkField := vModel.GetPrimaryField()
	if pkField == nil || pkField.GetValue() == nil || !pkField.GetValue().IsValid() || pkField.GetValue().IsZero() {
		err = cd.NewError(cd.IllegalParam, fmt.Sprintf("model %s missing primary key", vModel.GetPkgKey()))
		return
	}

	ret, err = s.modelCodec.PackedBasicFieldValue(pkField, pkField.GetValue())
	return
}

func buildContainPrimaryEntryMap(entries []containModelEntry) (ret map[string]containModelEntry, ok bool) {
	ret = make(map[string]containModelEntry, len(entries))
	for _, entry := range entries {
		if !entry.hasPrimary {
			return nil, false
		}
		if _, found := ret[entry.primaryID]; found {
			return nil, false
		}
		ret[entry.primaryID] = entry
	}
	ok = true
	return
}

func buildContainPrimaryEntryMapForUpdate(entries []containModelEntry) (ret map[string]containModelEntry, pendingInsert []containModelEntry, ok bool) {
	ret = make(map[string]containModelEntry, len(entries))
	pendingInsert = make([]containModelEntry, 0)
	for _, entry := range entries {
		if !entry.hasPrimary {
			pendingInsert = append(pendingInsert, entry)
			continue
		}
		if _, found := ret[entry.primaryID]; found {
			return nil, nil, false
		}
		ret[entry.primaryID] = entry
	}
	ok = true
	return
}
