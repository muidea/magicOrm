package models_test

import (
	"testing"

	"github.com/muidea/magicOrm/models"
)

func TestReadQueryKeysetLexicographic(t *testing.T) {
	order := []models.QueryOrder{{Field: "rank", Direction: "desc"}, {Field: "id", Direction: "asc"}}
	after := []models.QueryScalar{{Kind: "integer", Number: "20"}, {Kind: "integer", Number: "9"}}
	e, err := models.QueryKeyset(order, after)
	if err != nil {
		t.Fatal(err)
	}
	if e.Kind != "or" || len(e.Children) != 2 || e.Children[0].Operator != "lt" || e.Children[0].Field != "rank" {
		t.Fatal("降序首项边界错误", e)
	}
	second := e.Children[1]
	if second.Kind != "and" || len(second.Children) != 2 || second.Children[0].Operator != "eq" || second.Children[0].Field != "rank" || second.Children[1].Operator != "gt" || second.Children[1].Field != "id" {
		t.Fatal("相等前缀及升序末项错误", second)
	}
	after[0].Number = "changed"
	order[0].Field = "changed"
	if e.Children[0].Value.Number != "20" || e.Children[0].Field != "rank" {
		t.Fatal("keyset共享输入")
	}
	for _, n := range []int{0, 1, 8, 9} {
		orders := make([]models.QueryOrder, n)
		values := make([]models.QueryScalar, n)
		for i := range orders {
			orders[i] = models.QueryOrder{Field: string(rune('a' + i)), Direction: "asc"}
			values[i] = models.QueryScalar{Kind: "integer", Number: "0"}
		}
		_, err := models.QueryKeyset(orders, values)
		if (err == nil) != (n == 1 || n == 8) {
			t.Fatal("排序项预算错误", n, err)
		}
	}
}

func TestReadQueryPlanNodeBudget(t *testing.T) {
	leaf := models.QueryExpression{Kind: "compare", Field: "id", Operator: "gt", Value: &models.QueryScalar{Kind: "integer", Number: "0"}}
	for _, count := range []int{56, 64} {
		root := models.QueryExpression{Kind: "or"}
		for i := 0; i < count/8; i++ {
			group := models.QueryExpression{Kind: "and"}
			for j := 0; j < 8; j++ {
				group.Children = append(group.Children, leaf)
			}
			root.Children = append(root.Children, group)
		}
		err := models.ValidateReadQueryPlan(models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, Where: &root})
		if (err == nil) != (count == 56) {
			t.Fatal("总节点预算未执行", count, err)
		}
	}
}
