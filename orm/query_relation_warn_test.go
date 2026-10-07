package orm

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/provider"
)

type diagnosticRelationOwner struct {
	ID       int64                   `orm:"id key"`
	Secret   string                  `orm:"secret"`
	Child    queryMaskRelationChild  `orm:"child"`
	Optional *queryMaskRelationChild `orm:"optional"`
}

type diagnosticStringRelationOwner struct {
	ID    string                 `orm:"id key"`
	Child queryMaskRelationChild `orm:"child"`
}

func TestEmptySingleRelationDiagnosticsIdentifyOwnerWithoutChangingQueryResult(t *testing.T) {
	p := provider.NewLocalProvider("relation-diagnostics", nil)
	for _, value := range []any{&queryMaskRelationChild{}, &diagnosticRelationOwner{}} {
		if _, err := p.RegisterModel(value); err != nil {
			t.Fatal(err)
		}
	}
	owner, err := p.GetEntityModel(&diagnosticRelationOwner{ID: 42, Secret: "business-secret"}, true)
	if err != nil {
		t.Fatal(err)
	}
	runner := &QueryRunner{relationEdges: map[string][]any{}, baseRunner: baseRunner{modelCodec: codec.New(p, "tenant")}}
	var logs bytes.Buffer
	logger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	defer slog.SetDefault(logger)
	for _, name := range []string{"child", "optional"} {
		runner.cacheRelationEdge(owner.GetPkgKey(), name, int64(42), nil)
		if err := runner.querySingleRelation(owner, owner.GetField(name), 0); err != nil {
			t.Fatal("empty relation semantics changed:", err)
		}
	}
	var record map[string]any
	if err := json.Unmarshal(logs.Bytes(), &record); err != nil {
		t.Fatal("optional relation warned or diagnostic invalid:", err)
	}
	if record["model"] != owner.GetPkgKey() || record["owner_id"] != float64(42) || record["field"] != "child" || record["relation_model"] != owner.GetField("child").GetType().GetPkgKey() || record["reason"] != "non_pointer_relation_ids_empty" {
		t.Fatalf("relation identity or cause missing: %v", record)
	}
	if strings.Contains(logs.String(), "business-secret") {
		t.Fatal("full object leaked")
	}
}

func TestRelationDiagnosticsHashStringPrimaryKeys(t *testing.T) {
	p := provider.NewLocalProvider("relation-string-diagnostics", nil)
	for _, value := range []any{&queryMaskRelationChild{}, &diagnosticStringRelationOwner{}} {
		if _, err := p.RegisterModel(value); err != nil {
			t.Fatal(err)
		}
	}
	owner, err := p.GetEntityModel(&diagnosticStringRelationOwner{ID: "private-string-key"}, true)
	if err != nil {
		t.Fatal(err)
	}
	attrs := relationDiagnosticAttrs(owner, owner.GetField("child"))
	record := map[string]any{}
	for i := 0; i < len(attrs); i += 2 {
		record[attrs[i].(string)] = attrs[i+1]
	}
	digest := sha256.Sum256([]byte("private-string-key"))
	if record["owner_id_sha256"] != fmt.Sprintf("%x", digest) {
		t.Fatal("string identity missing")
	}
	if strings.Contains(fmt.Sprint(attrs), "private-string-key") {
		t.Fatal("raw string key leaked")
	}
}

func TestQueryRunnerShouldWarnRelationMissOncePerRelation(t *testing.T) {
	runner := &QueryRunner{
		relationWarns: map[string]struct{}{},
	}

	relationMissWarnTracker.Lock()
	origWarns := relationMissWarnTracker.lastWarnAt
	relationMissWarnTracker.lastWarnAt = map[string]time.Time{}
	relationMissWarnTracker.Unlock()
	defer func() {
		relationMissWarnTracker.Lock()
		relationMissWarnTracker.lastWarnAt = origWarns
		relationMissWarnTracker.Unlock()
	}()

	if !runner.shouldWarnRelationMiss("/vmi/product", int64(63)) {
		t.Fatal("first relation miss should emit warning")
	}
	if runner.shouldWarnRelationMiss("/vmi/product", int64(63)) {
		t.Fatal("duplicate relation miss should be suppressed")
	}
	if !runner.shouldWarnRelationMiss("/vmi/product", int64(64)) {
		t.Fatal("different relation id should emit warning")
	}
	if !runner.shouldWarnRelationMiss("/vmi/store", int64(63)) {
		t.Fatal("different relation model should emit warning")
	}
}
