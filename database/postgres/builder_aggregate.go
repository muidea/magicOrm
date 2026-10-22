package postgres

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
)

func (s *Builder) BuildAggregate(model models.Model, filter models.Filter, plan models.AggregatePlan) (database.Result, *cd.Error) {
	projection, groups, projectionArgs, err := database.BuildAggregateProjection(model, plan, "postgres", s.modelProvider, 0)
	if err != nil {
		return nil, err
	}
	if err = database.ValidateAggregateFilter(filter); err != nil {
		return nil, err
	}
	result := &ResultStack{}
	result.PushArgs(projectionArgs...)
	where, err := s.buildFilter(model, filter, result)
	if err != nil {
		return nil, err
	}
	planWhere, _, args, err := database.BuildReadQueryClause(model, filter, s.modelProvider, "postgres", len(result.Args()))
	if err != nil {
		return nil, err
	}
	result.PushArgs(args...)
	if planWhere != "" {
		if where == "" {
			where = planWhere
		} else {
			where = "(" + where + ") AND (" + planWhere + ")"
		}
	}
	bucketWhere, bucketArgs, err := database.BuildAggregateBucketClause(model, plan.Bucket, s.modelProvider, "postgres", len(result.Args()))
	if err != nil {
		return nil, err
	}
	result.PushArgs(bucketArgs...)
	if bucketWhere != "" {
		if where == "" {
			where = bucketWhere
		} else {
			where = "(" + where + ") AND (" + bucketWhere + ")"
		}
	}
	query := fmt.Sprintf("SELECT %s FROM \"%s\"", projection, s.buildCodec.ConstructModelTableName(model))
	if where != "" {
		query += " WHERE " + where
	}
	if groups != "" {
		query += " GROUP BY " + groups + " ORDER BY " + groups
	}
	query += fmt.Sprintf(" LIMIT $%d", len(result.Args())+1)
	result.PushArgs(plan.MaxGroups + 1)
	result.SetSQL(query)
	return result, nil
}
