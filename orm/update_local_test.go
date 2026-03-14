package orm

import (
	"context"
	"testing"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/local"
)

type localNilRelationChild struct {
	ID   int    `orm:"id key auto"`
	Name string `orm:"name"`
}

type localNilRelationParent struct {
	ID    int                    `orm:"id key auto"`
	Name  string                 `orm:"name"`
	Child *localNilRelationChild `orm:"child"`
}

func TestCompareRelationSingleFieldValueHandlesTypedNil(t *testing.T) {
	localProvider := provider.NewLocalProvider("tenant", nil)
	runner := NewUpdateRunner(context.Background(), nil, &fakeExecutor{}, localProvider, codec.New(localProvider, "tenant"))

	existingModel, err := local.GetEntityModel(&localNilRelationParent{ID: 1, Name: "left"}, nil)
	if err != nil {
		t.Fatalf("GetEntityModel(existing) failed: %v", err)
	}
	newModel, err := local.GetEntityModel(&localNilRelationParent{ID: 1, Name: "right"}, nil)
	if err != nil {
		t.Fatalf("GetEntityModel(new) failed: %v", err)
	}

	same, err := runner.compareRelationSingleFieldValue(existingModel.GetField("child"), newModel.GetField("child"), 0)
	if err != nil {
		t.Fatalf("compareRelationSingleFieldValue failed: %v", err)
	}
	if !same {
		t.Fatal("typed nil relations should compare as equal")
	}
}
