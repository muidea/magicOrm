package orm

import (
	"fmt"
	"log/slog"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/models"
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
	newVal := vField.GetValue().Get()
	err = s.deleteRelation(vModel, vField, 0)
	if err != nil {
		slog.Error("UpdateRunner updateContainRelation deleteRelation failed", "field", vField.GetName(), "error", err.Error())
		return
	}
	vField.SetValue(newVal)
	err = s.insertRelation(vModel, vField)
	if err != nil {
		slog.Error("UpdateRunner updateContainRelation insertRelation failed", "field", vField.GetName(), "error", err.Error())
	}
	return
}
