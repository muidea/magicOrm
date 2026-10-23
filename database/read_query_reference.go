package database

import (
	"fmt"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

// Semi-joins preserve row and aggregate cardinality for either reference storage.
// No target IDs are materialized in the caller, and no caller supplies table names.
func buildReadReferences(model models.Model, refs []models.QueryReference, p provider.Provider, dialect string, offset int, c codec.Codec, depth int) (string, []any, *cd.Error) {
	invalid := func() (string, []any, *cd.Error) {
		return "", nil, cd.NewError(cd.IllegalParam, "reference query does not match registered readable single references")
	}
	quote := func(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }
	if dialect == "mysql" {
		quote = func(s string) string { return "`" + strings.ReplaceAll(s, "`", "``") + "`" }
	}
	clauses, args := []string{}, []any{}
	for _, r := range refs {
		f := model.GetField(r.Field)
		if !readQueryIdentifier.MatchString(r.Field) || !models.IsSingleReferenceField(f) || f.GetSpec() == nil {
			return invalid()
		}
		if x := f.GetSpec().GetConstraints(); x != nil && x.Has(models.KeyWriteOnly) {
			return invalid()
		}
		target, err := p.GetTypeModel(f.GetType())
		if err != nil {
			return "", nil, err
		}
		if target == nil || target.GetPrimaryField() == nil {
			return invalid()
		}
		primary, err := AggregateField(target, target.GetPrimaryField().GetName())
		if err != nil || primary.GetType().IsPtrType() {
			return invalid()
		}
		base, err := p.GetModelFilter(target)
		if err != nil {
			return "", nil, err
		}
		filter, err := models.WithReadQueryPlan(base, models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, Where: r.Where, References: r.References})
		if err != nil {
			return "", nil, err
		}
		where, _, values, err := buildReadQueryClause(target, filter, p, dialect, offset+len(args), c, depth+1)
		if err != nil {
			return "", nil, err
		}
		args = append(args, values...)
		if r.Values != nil {
			keys := make([]string, 0, len(r.Values))
			for _, v := range r.Values {
				value, e := readQueryScalarValue(v, primary.GetType())
				if e != nil {
					return "", nil, e
				}
				encoded, e := p.EncodeValue(value, primary.GetType())
				if e != nil {
					return "", nil, e
				}
				args = append(args, encoded)
				placeholder := "?"
				if dialect == "postgres" {
					placeholder = fmt.Sprintf("$%d", offset+len(args))
				}
				keys = append(keys, placeholder)
			}
			keyWhere := "1=0"
			if len(keys) > 0 {
				keyWhere = quote(primary.GetName()) + " IN (" + strings.Join(keys, ",") + ")"
			}
			if where != "" {
				where = "(" + where + ") AND (" + keyWhere + ")"
			} else {
				where = keyWhere
			}
		}
		sub := "SELECT " + quote(primary.GetName()) + " FROM " + quote(c.ConstructModelTableName(target))
		if where != "" {
			sub += " WHERE " + where
		}
		if models.IsInlineReference(model, f) {
			clauses = append(clauses, quote(r.Field)+" IN ("+sub+")")
			continue
		}
		relation, err := c.ConstructRelationTableName(model, f)
		if err != nil {
			return "", nil, err
		}
		left := model.GetPrimaryField()
		if left == nil {
			return invalid()
		}
		clauses = append(clauses, quote(left.GetName())+" IN (SELECT "+quote("left")+" FROM "+quote(relation)+" WHERE "+quote("right")+" IN ("+sub+"))")
	}
	return strings.Join(clauses, " AND "), args, nil
}
