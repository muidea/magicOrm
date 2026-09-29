package orm

import (
	"testing"

	cd "github.com/muidea/magicCommon/def"
)

func TestFinalTransactionReturnsCommitError(t *testing.T) {
	executor := &fakeExecutor{commitErr: cd.NewError(cd.Unexpected, "commit unavailable")}
	ormImpl := &impl{executor: executor}
	if err := ormImpl.finalTransaction(nil); err == nil || err.Message != "commit unavailable" {
		t.Fatalf("commit failure was ignored: %v", err)
	}
	if executor.commitCalls != 1 {
		t.Fatalf("expected one commit, got %d", executor.commitCalls)
	}
}

func TestFinalTransactionUsesFinalErrorState(t *testing.T) {
	executor := &fakeExecutor{}
	ormImpl := &impl{executor: executor}

	func() {
		var err *cd.Error
		if beginErr := ormImpl.executor.BeginTransaction(); beginErr != nil {
			t.Fatalf("BeginTransaction failed: %v", beginErr)
		}
		defer func() {
			ormImpl.finalTransaction(err)
		}()

		err = cd.NewError(cd.Unexpected, "boom")
	}()

	if executor.commitCalls != 0 {
		t.Fatalf("expected no commit on error, got %d", executor.commitCalls)
	}
	if executor.rollbackCalls != 1 {
		t.Fatalf("expected rollback on error, got %d", executor.rollbackCalls)
	}
}

func TestFinalTransactionCommitsWhenNoError(t *testing.T) {
	executor := &fakeExecutor{}
	ormImpl := &impl{executor: executor}

	func() {
		var err *cd.Error
		if beginErr := ormImpl.executor.BeginTransaction(); beginErr != nil {
			t.Fatalf("BeginTransaction failed: %v", beginErr)
		}
		defer func() {
			ormImpl.finalTransaction(err)
		}()
	}()

	if executor.commitCalls != 1 {
		t.Fatalf("expected commit without error, got %d", executor.commitCalls)
	}
	if executor.rollbackCalls != 0 {
		t.Fatalf("expected no rollback without error, got %d", executor.rollbackCalls)
	}
}
