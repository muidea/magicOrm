package database_test

import (
	"strings"
	"testing"

	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/internal/database/mysql"
	"github.com/muidea/magicOrm/internal/database/postgres"
	"github.com/muidea/magicOrm/models"
)

func TestAggregateReferenceGroupsPreserveRootCardinality(t *testing.T) {
	for _, useRemote := range []bool{false, true} {
		p, event := referenceQueryModel(t, useRemote)
		c := codec.New(p, "isolated")
		asset, e := p.GetTypeModel(event.GetField("asset").GetType())
		if e != nil {
			t.Fatal(e)
		}
		for dialect, builder := range map[string]database.Builder{"postgres": postgres.NewBuilder(p, c), "mysql": mysql.NewBuilder(p, c)} {
			for _, tc := range []struct {
				model  models.Model
				field  string
				inline bool
			}{{event, "asset", false}, {asset, "zone", true}} {
				f, _ := p.GetModelFilter(tc.model)
				result, e := builder.(database.AggregateBuilder).BuildAggregate(tc.model, f, models.AggregatePlan{APIVersion: models.AggregatePlanAPIVersion, GroupBy: []string{tc.field}, Metrics: []models.AggregateMetric{{Key: "n", Operation: "count"}}, MaxGroups: 1000})
				if e != nil {
					t.Fatal(e)
				}
				sql := result.SQL()
				quote := "\""
				if dialect == "mysql" {
					quote = "`"
				}
				if strings.Contains(sql, "JOIN") || !strings.Contains(sql, "GROUP BY 1 ORDER BY 1") {
					t.Fatal(sql)
				}
				if tc.inline {
					if !strings.HasPrefix(sql, "SELECT "+quote+tc.field+quote+", COUNT(*)") {
						t.Fatal(sql)
					}
				} else {
					if !strings.Contains(sql, "SELECT "+quote+"right"+quote+" FROM ") {
						t.Fatal(sql)
					}
				}
				scalar, e := database.AggregateGroupField(tc.model, tc.field, p)
				if e != nil || scalar.GetName() != "id" || scalar.GetType().GetValue() != models.TypeBigIntegerValue {
					t.Fatal(scalar, e)
				}
			}
			for _, field := range []string{"secret", "many", "asset)"} {
				f, _ := p.GetModelFilter(event)
				if _, e := builder.(database.AggregateBuilder).BuildAggregate(event, f, models.AggregatePlan{APIVersion: models.AggregatePlanAPIVersion, GroupBy: []string{field}, Metrics: []models.AggregateMetric{{Key: "n", Operation: "count"}}, MaxGroups: 10}); e == nil {
					t.Fatal("invalid reference grouping accepted", field)
				}
			}
		}
	}
}
