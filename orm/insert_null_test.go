package orm

import (
	"context"
	"strings"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

type NullInsertTarget struct {
	ID int64 `orm:"id key" view:"detail,lite"`
}
type NullInsertHost struct {
	ID  int64             `orm:"id key" view:"detail,lite"`
	Ref *NullInsertTarget `orm:"ref" view:"detail,lite"`
}
type NullInsertRequiredHost struct {
	ID  int64             `orm:"id key" view:"detail,lite"`
	Ref *NullInsertTarget `orm:"ref" constraint:"req" view:"detail,lite"`
}

func nullInsertModel(t *testing.T, isRemote, required bool) (provider.Provider, models.Model) {
	t.Helper()
	p := provider.NewLocalProvider("null-insert", nil)
	if isRemote {
		p = provider.NewRemoteProvider("null-insert", nil)
	}
	var host any = &NullInsertHost{ID: 1}
	if required {
		host = &NullInsertRequiredHost{ID: 1}
	}
	for _, entity := range []any{&NullInsertTarget{}, host} {
		var definition any = entity
		if isRemote {
			object, err := helper.GetObject(entity)
			if err != nil {
				t.Fatal(err)
			}
			definition = object
		}
		if _, err := p.RegisterModel(definition); err != nil {
			t.Fatal(err)
		}
	}
	if isRemote {
		value, err := helper.GetObjectValue(host)
		if err != nil {
			t.Fatal(err)
		}
		// Omitted pointer fields must remain unassigned, as in a JSON body
		// that has only the primary key.
		value.Fields = []*remote.FieldValue{{Name: "id", Value: int64(1), Assigned: true}}
		host = value
	}
	m, err := p.GetEntityModel(host, true)
	if err != nil {
		t.Fatal(err)
	}
	return p, m
}

func TestInsertOptionalAndRequiredReferencePresence(t *testing.T) {
	for _, isRemote := range []bool{false, true} {
		for _, required := range []bool{false, true} {
			for _, mode := range []string{"omitted", "null", "value"} {
				name := "local/optional/" + mode
				if isRemote {
					name = "remote/optional/" + mode
				}
				if required {
					name = strings.Replace(name, "optional", "required", 1)
				}
				t.Run(name, func(t *testing.T) {
					p, m := nullInsertModel(t, isRemote, required)
					if mode == "null" {
						if err := m.SetFieldValue("ref", nil); err != nil {
							t.Fatal(err)
						}
					}
					if mode == "value" {
						var value any = &NullInsertTarget{ID: 2}
						if isRemote {
							var err *cd.Error
							value, err = helper.GetObjectValue(value)
							if err != nil {
								t.Fatal(err)
							}
						}
						if err := m.SetFieldValue("ref", value); err != nil {
							t.Fatal(err)
						}
					}
					executor := &fakeExecutor{}
					impl := &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: codec.New(p, "test"), validationMgr: newDefaultValidationManager()}
					_, err := impl.Insert(m)
					if required && mode != "value" {
						if err == nil || !strings.Contains(err.Message, "ref") {
							t.Fatalf("required reference admitted: %v", err)
						}
						if len(executor.execCalls) != 0 || executor.beginCalls != 0 {
							t.Fatal("invalid relation started a write transaction")
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					want := 1
					if mode == "value" {
						want = 2
					}
					if len(executor.execCalls) != want || executor.commitCalls != 1 || executor.rollbackCalls != 0 {
						t.Fatalf("unexpected writes: %+v", executor)
					}
					if mode == "null" && !models.IsAssignedField(m.GetField("ref")) {
						t.Fatal("explicit clear assignment was lost")
					}
				})
			}
		}
	}
}

type panickingInsertExecutor struct{ fakeExecutor }

func (e *panickingInsertExecutor) ExecuteInsert(string, any, ...any) *cd.Error {
	panic("executor fault")
}

func TestInsertPanicRollsBackAndPreservesPanic(t *testing.T) {
	p, m := nullInsertModel(t, true, false)
	executor := &panickingInsertExecutor{}
	handler := &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: codec.New(p, "test"), validationMgr: newDefaultValidationManager()}
	func() {
		defer func() {
			if got := recover(); got != "executor fault" {
				t.Fatalf("unexpected panic: %v", got)
			}
		}()
		_, _ = handler.Insert(m)
	}()
	if executor.commitCalls != 0 || executor.rollbackCalls != 1 {
		t.Fatalf("panic transaction finalized incorrectly: %+v", executor)
	}
}

type failingRelationExecutor struct{ fakeExecutor }

func (e *failingRelationExecutor) ExecuteInsert(sql string, out any, args ...any) *cd.Error {
	if len(e.execCalls) > 0 {
		return cd.NewError(cd.Unexpected, "relation write failure")
	}
	return e.fakeExecutor.ExecuteInsert(sql, out, args...)
}

func TestRelationWriteFailureRollsBackHost(t *testing.T) {
	p, m := nullInsertModel(t, true, false)
	value, err := helper.GetObjectValue(&NullInsertTarget{ID: 2})
	if err != nil {
		t.Fatal(err)
	}
	if err = m.SetFieldValue("ref", value); err != nil {
		t.Fatal(err)
	}
	executor := &failingRelationExecutor{}
	handler := &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: codec.New(p, "test"), validationMgr: newDefaultValidationManager()}
	if _, err = handler.Insert(m); err == nil {
		t.Fatal("relation write error ignored")
	}
	if executor.commitCalls != 0 || executor.rollbackCalls != 1 {
		t.Fatalf("failed relation committed host: %+v", executor)
	}
}

func TestAbsentReferenceClearsQuerySelectionSeed(t *testing.T) {
	for _, isRemote := range []bool{false, true} {
		p, m := nullInsertModel(t, isRemote, false)
		var seed any = &NullInsertTarget{ID: 2}
		if isRemote {
			var err *cd.Error
			seed, err = helper.GetObjectValue(seed)
			if err != nil {
				t.Fatal(err)
			}
		}
		if err := m.SetFieldValue("ref", seed); err != nil {
			t.Fatal(err)
		}
		runner := &QueryRunner{relationEdges: map[string][]any{}, baseRunner: baseRunner{modelProvider: p, modelCodec: codec.New(p, "test")}}
		runner.cacheRelationEdge(m.GetPkgKey(), "ref", int64(1), nil)
		if err := runner.querySingleRelation(m, m.GetField("ref"), 0); err != nil {
			t.Fatal(err)
		}
		if models.IsValidField(m.GetField("ref")) {
			t.Fatalf("remote=%v: query invented reference", isRemote)
		}
	}
}
