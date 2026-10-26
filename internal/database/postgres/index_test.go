package postgres

import (
	"strings"
	"testing"

	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

func TestBuildCreateTableIncludesDeclaredIndex(t *testing.T) {
	object, err := helper.GetObject(&postgresUniqueModel{})
	if err != nil {
		t.Fatal(err)
	}
	object.Indexes = []models.Index{{Name: "idx_namespace_account", Fields: []string{"namespace", "accountId"}}}
	remoteProvider := provider.NewRemoteProvider("test", nil)
	model, err := remoteProvider.RegisterModel(object)
	if err != nil {
		t.Fatal(err)
	}
	builder := NewBuilder(remoteProvider, codec.New(remoteProvider, "cas"))
	result, buildErr := builder.BuildCreateTable(model)
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	if !strings.Contains(result.SQL(), `CREATE INDEX IF NOT EXISTS "cas_PostgresUniqueModel_idx_namespace_account_d6d25cd7ee5d6942" ON "cas_PostgresUniqueModel" ("namespace", "accountId")`) {
		t.Fatalf("index missing from SQL: %s", result.SQL())
	}
}

func TestIndexBuilderRejectsRelationColumnButAllowsInlineReference(t *testing.T) {
	for _, inline := range []bool{false, true} {
		p := provider.NewRemoteProvider("index-storage", nil)
		target, err := helper.GetObject(&physicalKey{})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = p.RegisterModel(target); err != nil {
			t.Fatal(err)
		}
		object, err := helper.GetObject(&indexReferenceRecord{})
		if err != nil {
			t.Fatal(err)
		}
		object.Indexes = []models.Index{{Name: "idx_related", Fields: []string{"related", "name"}}}
		if inline {
			object.UniqueConstraints = []models.UniqueConstraint{{Name: "uq_related", Fields: []string{"related", "name"}}}
		}
		model, err := p.RegisterModel(object)
		if err != nil {
			t.Fatal("old declarations must remain inspectable", err)
		}
		builder := NewBuilder(p, codec.New(p, "index"))
		_, err = builder.BuildCreateTable(model)
		if (err == nil) != inline {
			t.Fatal(inline, err)
		}
		_, err = builder.BuildCreateIndex(model, object.Indexes[0])
		if (err == nil) != inline {
			t.Fatal(inline, err)
		}
		if _, err = builder.DescribeTable(model); err != nil {
			t.Fatal("read-only legacy description blocked", err)
		}
	}
}

type indexReferenceRecord struct {
	ID      int64        `orm:"id key"`
	Name    string       `orm:"name"`
	Related *physicalKey `orm:"related"`
}
