package postgres

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"

	"log/slog"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
)

// BuildQuery build query sql
func (s *Builder) BuildQuery(vModel models.Model, filter models.Filter) (ret database.Result, err *cd.Error) {
	namesVal, nameErr := s.getFieldQueryNames(vModel)
	if nameErr != nil {
		err = nameErr
		slog.Error("BuildQuery failed", "operation", "s.getFieldQueryNames", "error", err.Error())
		return
	}

	resultStackPtr := &ResultStack{}
	querySQL := fmt.Sprintf("SELECT %s FROM \"%s\"", namesVal, s.buildCodec.ConstructModelTableName(vModel))
	if filter != nil {
		filterSQL, filterErr := s.buildFilter(vModel, filter, resultStackPtr)
		if filterErr != nil {
			err = filterErr
			slog.Error("BuildQuery failed", "operation", "s.buildFilter", "error", err.Error())
			return
		}

		if filterSQL != "" {
			querySQL = fmt.Sprintf("%s WHERE %s", querySQL, filterSQL)
		}

		sortVal, sortErr := s.buildSorter(vModel, filter.Sorter())
		if sortErr != nil {
			err = sortErr
			slog.Error("BuildQuery failed", "operation", "s.buildSorter", "error", err.Error())
			return
		}

		if sortVal != "" {
			querySQL = fmt.Sprintf("%s ORDER BY %s", querySQL, sortVal)
		}

		paginationer := filter.Paginationer()
		if paginationer != nil {
			resultStackPtr.PushArgs(paginationer.Limit(), paginationer.Offset())
			querySQL = fmt.Sprintf("%s LIMIT $%d OFFSET $%d", querySQL, len(resultStackPtr.argsVal)-1, len(resultStackPtr.argsVal))
		}
	}
	if traceSQL() {
		slog.Info("[SQL] query", "sql", querySQL)
	}

	resultStackPtr.SetSQL(querySQL)
	ret = resultStackPtr
	return
}

// BuildQueryRelation build query relation sql
func (s *Builder) BuildQueryRelation(vModel models.Model, vField models.Field) (ret database.Result, err *cd.Error) {
	leftVal := vModel.GetPrimaryField().GetValue().Get()
	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField)
	if relationErr != nil {
		err = relationErr
		slog.Error("BuildQueryRelation failed", "field", vField.GetName(), "operation", "s.buildCodec.ConstructRelationTableName", "error", err.Error())
		return
	}

	resultStackPtr := &ResultStack{}
	resultStackPtr.PushArgs(leftVal)
	queryRelationSQL := fmt.Sprintf("SELECT \"right\" FROM \"%s\" WHERE \"left\"= $%d", relationTableName, len(resultStackPtr.argsVal))
	if traceSQL() {
		slog.Info("[SQL] query relation", "sql", queryRelationSQL)
	}

	resultStackPtr.SetSQL(queryRelationSQL)
	ret = resultStackPtr
	return
}

func (s *Builder) BuildBatchQueryRelation(vModel models.Model, vField models.Field, leftIDs []any) (ret database.Result, err *cd.Error) {
	if len(leftIDs) == 0 {
		err = cd.NewError(cd.IllegalParam, "leftIDs is empty")
		return
	}

	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField)
	if relationErr != nil {
		err = relationErr
		slog.Error("BuildBatchQueryRelation failed", "field", vField.GetName(), "operation", "s.buildCodec.ConstructRelationTableName", "error", err.Error())
		return
	}

	pkField := vModel.GetPrimaryField()
	if pkField == nil {
		err = cd.NewError(cd.IllegalParam, "model primary field is nil")
		return
	}

	resultStackPtr := &ResultStack{}
	placeholders := ""
	for _, leftID := range leftIDs {
		encodedVal, encodeErr := s.modelProvider.EncodeValue(leftID, pkField.GetType())
		if encodeErr != nil {
			err = encodeErr
			slog.Error("BuildBatchQueryRelation failed", "field", vField.GetName(), "operation", "s.modelProvider.EncodeValue", "error", err.Error())
			return
		}

		resultStackPtr.PushArgs(encodedVal)
		if placeholders == "" {
			placeholders = fmt.Sprintf("$%d", len(resultStackPtr.argsVal))
		} else {
			placeholders = fmt.Sprintf("%s,$%d", placeholders, len(resultStackPtr.argsVal))
		}
	}

	queryRelationSQL := fmt.Sprintf("SELECT \"left\",\"right\" FROM \"%s\" WHERE \"left\" IN (%s)", relationTableName, placeholders)
	if traceSQL() {
		slog.Info("[SQL] batch query relation", "sql", queryRelationSQL)
	}

	resultStackPtr.SetSQL(queryRelationSQL)
	ret = resultStackPtr
	return
}

func (s *Builder) getFieldQueryNames(vModel models.Model) (ret string, err *cd.Error) {
	str := ""
	for _, field := range vModel.GetFields() {
		// 检查 wo 约束，这些字段在查询时应该被排除
		fSpec := field.GetSpec()
		constraints := fSpec.GetConstraints()
		if constraints != nil && constraints.Has(models.KeyWriteOnly) {
			continue
		}
		// Query 时：基础列中，已赋值或“值类型 slice”（如 []int）均拉取，以便完整加载行；指针型未赋值不拉取
		if !models.IsBasicField(field) {
			continue
		}
		if !models.IsValidField(field) && !(models.IsSliceField(field) && !models.IsPtrField(field)) {
			continue
		}

		if str == "" {
			str = fmt.Sprintf("\"%s\"", field.GetName())
		} else {
			str = fmt.Sprintf("%s,\"%s\"", str, field.GetName())
		}
	}

	ret = str
	return
}

func (s *Builder) BuildModuleValueHolder(vModel models.Model) (ret []any, err *cd.Error) {
	items := []any{}
	for _, field := range vModel.GetFields() {
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

		itemVal, itemErr := getFieldValueHolder(field.GetType())
		if itemErr != nil {
			err = itemErr
			slog.Error("BuildModuleValueHolder failed", "operation", "getFieldPlaceHolder", "error", err.Error())
			return
		}

		items = append(items, itemVal)
	}

	ret = items
	return
}
