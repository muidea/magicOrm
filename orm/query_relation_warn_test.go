package orm

import (
	"testing"
	"time"
)

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
