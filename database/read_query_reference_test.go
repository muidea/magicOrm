package database_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/database/mysql"
	"github.com/muidea/magicOrm/database/postgres"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

type scopeZone struct {
	ID     int64  `orm:"id key" view:"detail,lite"`
	Name   string `orm:"name" view:"detail,lite"`
	Secret string `orm:"secret" constraint:"wo"`
}
type scopeAsset struct {
	ID    int64      `orm:"id key" view:"detail,lite"`
	State string     `orm:"state" view:"detail,lite"`
	Zone  *scopeZone `orm:"zone" view:"detail,lite"`
}

func (*scopeAsset) GetUniqueConstraints() []models.UniqueConstraint {
	return []models.UniqueConstraint{{Name: "asset_zone", Fields: []string{"zone", "id"}}}
}

type scopeEvent struct {
	ID     int64         `orm:"id key" view:"detail,lite"`
	Rank   int64         `orm:"rank" view:"detail,lite"`
	Asset  *scopeAsset   `orm:"asset" view:"detail,lite"`
	Secret *scopeAsset   `orm:"secret" constraint:"wo"`
	Many   []*scopeAsset `orm:"many" view:"detail,lite"`
}

func referenceQueryModel(t *testing.T, useRemote bool) (provider.Provider, models.Model) {
	t.Helper()
	p := provider.NewLocalProvider("scope", nil)
	if useRemote {
		p = provider.NewRemoteProvider("scope", nil)
	}
	var model models.Model
	for _, value := range []any{&scopeZone{}, &scopeAsset{}, &scopeEvent{}} {
		if useRemote {
			obj, err := helper.GetObject(value)
			if err != nil {
				t.Fatal(err)
			}
			value = obj
		}
		m, registerErr := p.RegisterModel(value)
		if registerErr != nil {
			t.Fatal(registerErr)
		}
		model = m
	}
	return p, model
}

func TestReadQueryReferenceSemiJoin(t *testing.T) {
	for _, useRemote := range []bool{false, true} {
		p, model := referenceQueryModel(t, useRemote)
		c := codec.New(p, "isolated")
		for dialect, builder := range map[string]database.Builder{"postgres": postgres.NewBuilder(p, c), "mysql": mysql.NewBuilder(p, c)} {
			t.Run(dialect+map[bool]string{false: "/local", true: "/remote"}[useRemote], func(t *testing.T) {
				base, e := p.GetModelFilter(model)
				if e != nil {
					t.Fatal(e)
				}
				base.Pagination(1, 21)
				if e := base.Equal("rank", int64(3)); e != nil {
					t.Fatal(e)
				}
				root := queryCompare("rank", "gt", "integer", "4")
				target := queryCompare("state", "eq", "string", "x' OR 1=1 --")
				zone := queryCompare("name", "eq", "string", "north")
				plan := models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, Where: &root, References: []models.QueryReference{{Field: "asset", Where: &target, Values: []models.QueryScalar{{Kind: "integer", Number: "9223372036854775001"}}, References: []models.QueryReference{{Field: "zone", Where: &zone}}}}}
				filter, e := models.WithReadQueryPlan(base, plan)
				if e != nil {
					t.Fatal(e)
				}
				// Both caller input and readback are isolated, including nested scalars.
				plan.References[0].Where.Value.Text = "changed"
				copy := filter.(models.ReadQueryPlanProvider).ReadQueryPlan()
				copy.References[0].References[0].Where.Value.Text = "changed"
				count, e := builder.BuildCount(model, filter)
				if e != nil {
					t.Fatal(e)
				}
				sql := count.SQL()
				quote := "\""
				if dialect == "mysql" {
					quote = "`"
				}
				q := func(s string) string { return quote + s + quote }
				relation, e := c.ConstructRelationTableName(model, model.GetField("asset"))
				if e != nil {
					t.Fatal(e)
				}
				for _, part := range []string{q("id") + " IN (SELECT " + q("left") + " FROM " + q(relation), "SELECT " + q("id") + " FROM " + q("isolated_ScopeAsset"), q("zone") + " IN (SELECT " + q("id") + " FROM " + q("isolated_ScopeZone")} {
					if !strings.Contains(sql, part) {
						t.Fatal(part, sql)
					}
				}
				if strings.Contains(sql, "JOIN") || strings.Contains(sql, "OR 1=1") || strings.Contains(sql, "9223372036854775001") {
					t.Fatal(sql)
				}
				want := []any{int64(3), int64(4), "x' OR 1=1 --", "north", int64(9223372036854775001)}
				if !reflect.DeepEqual(count.Args(), want) {
					t.Fatalf("args=%#v", count.Args())
				}
				if dialect == "postgres" && !strings.Contains(sql, "$5") {
					t.Fatal(sql)
				}
				result, e := builder.BuildQuery(model, filter)
				if e != nil {
					t.Fatal(e)
				}
				if !reflect.DeepEqual(result.Args()[:5], want) {
					t.Fatal(result.Args())
				}
				aggregateBase, _ := p.GetModelFilter(model)
				if e := aggregateBase.Equal("rank", int64(3)); e != nil {
					t.Fatal(e)
				}
				aggregateFilter, e := models.WithReadQueryPlan(aggregateBase, filter.(models.ReadQueryPlanProvider).ReadQueryPlan())
				if e != nil {
					t.Fatal(e)
				}
				ag, e := builder.(database.AggregateBuilder).BuildAggregate(model, aggregateFilter, models.AggregatePlan{APIVersion: models.AggregatePlanAPIVersion, Metrics: []models.AggregateMetric{{Key: "n", Operation: "count"}}, MaxGroups: 1})
				if e != nil {
					t.Fatal(e)
				}
				if !reflect.DeepEqual(ag.Args()[:5], want) || strings.Contains(ag.SQL(), "JOIN") {
					t.Fatal(ag.SQL(), ag.Args())
				}
				if _, e := builder.BuildUpdateWithFilter(model, filter); e == nil {
					t.Fatal("read plan used for write")
				}
			})
		}
	}
}

func TestReadQueryReferenceRejectsAndEmptySet(t *testing.T) {
	for _, useRemote := range []bool{false, true} {
		p, model := referenceQueryModel(t, useRemote)
		c := codec.New(p, "isolated")
		builder := postgres.NewBuilder(p, c)
		for _, field := range []string{"many", "secret", "unknown", "rank", "asset) OR 1=1"} {
			base, _ := p.GetModelFilter(model)
			f, e := models.WithReadQueryPlan(base, models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, References: []models.QueryReference{{Field: field}}})
			if e == nil {
				_, e = builder.BuildCount(model, f)
			}
			if e == nil {
				t.Fatal("accepted", field)
			}
		}
		for _, keys := range [][]models.QueryScalar{nil, {}, {{Kind: "integer", Number: "9223372036854775808"}}, {{Kind: "string", Text: "1"}}} {
			base, _ := p.GetModelFilter(model)
			plan := models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, References: []models.QueryReference{{Field: "asset", Values: keys}}}
			bytes, _ := json.Marshal(plan)
			var roundtrip models.ReadQueryPlan
			if e := json.Unmarshal(bytes, &roundtrip); e != nil {
				t.Fatal(e)
			}
			f, e := models.WithReadQueryPlan(base, roundtrip)
			if e != nil {
				t.Fatal(e)
			}
			result, e := builder.BuildCount(model, f)
			if len(keys) > 0 {
				if e == nil {
					t.Fatal("accepted invalid primary", keys)
				}
				continue
			}
			if e != nil {
				t.Fatal(e)
			}
			if strings.Contains(result.SQL(), "1=0") != (keys != nil) {
				t.Fatal("empty set lost", result.SQL())
			}
		}
		bad := queryCompare("secret", "eq", "string", "hidden")
		base, _ := p.GetModelFilter(model)
		f, _ := models.WithReadQueryPlan(base, models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, References: []models.QueryReference{{Field: "asset", References: []models.QueryReference{{Field: "zone", Where: &bad}}}}})
		if _, e := builder.BuildCount(model, f); e == nil {
			t.Fatal("write-only target used in predicate")
		}
	}
}

func TestReadQueryReferenceBudgets(t *testing.T) {
	for _, refs := range [][]models.QueryReference{
		{{Field: "a"}, {Field: "a"}},
		{{Field: "a", Values: []models.QueryScalar{{Kind: "integer", Number: "1"}, {Kind: "integer", Number: "1"}}}},
		{{Field: "a", References: []models.QueryReference{{Field: "b", References: []models.QueryReference{{Field: "c", References: []models.QueryReference{{Field: "d"}}}}}}}},
	} {
		if e := models.ValidateReadQueryPlan(models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, References: refs}); e == nil {
			t.Fatal("over budget plan accepted")
		}
	}
}
