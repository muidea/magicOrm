package orm

import (
	"testing"

	cd "github.com/muidea/magicCommon/def"
)

type serializableORMExecutor struct {
	fakeExecutor
	calls int
	err   *cd.Error
}

func (s *serializableORMExecutor) BeginSerializableTransaction() *cd.Error {
	s.calls++
	return s.err
}

func TestSerializableTransactionCapability(t *testing.T) {
	for _, handler := range []Orm{nil, (*impl)(nil), &impl{}, &impl{executor: &fakeExecutor{}}} {
		if BeginSerializableTransaction(handler) == nil {
			t.Fatal("缺失能力时未明确拒绝")
		}
	}
	executor := &serializableORMExecutor{}
	handler := &impl{executor: executor}
	if err := BeginSerializableTransaction(handler); err != nil || executor.calls != 1 {
		t.Fatal("未调用强隔离能力", err, executor.calls)
	}
	executor.err = cd.NewError(cd.InvalidOperation, "begin rejected")
	if err := BeginSerializableTransaction(handler); err != executor.err || executor.calls != 2 {
		t.Fatal("错误被吞掉或自动重试", err, executor.calls)
	}
}
