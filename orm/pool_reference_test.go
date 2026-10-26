package orm

import (
	"sync"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/internal/database"
)

type reviewPool struct {
	database.Pool
	refs   int
	closed bool
}

func (p *reviewPool) IncReference() int            { p.refs++; return p.refs }
func (p *reviewPool) DecReference() int            { p.refs--; return p.refs }
func (p *reviewPool) GetStats() database.PoolStats { return database.PoolStats{} }
func (p *reviewPool) CheckConfig(database.Config) *cd.Error {
	return cd.NewError(cd.IllegalParam, "configuration mismatch")
}
func (p *reviewPool) Uninitialized() { p.closed = true }
func TestRejectedDatabaseBindingKeepsReferenceCount(t *testing.T) {
	const owner = "temporary-capability-review"
	p := &reviewPool{refs: 1}
	name2Pool.Store(owner, p)
	defer name2Pool.Delete(owner)
	defer name2LastPoolStats.Delete(owner)
	if err := AddDatabase("unused", "unused", "public", "unused", "unused", 1, owner); err == nil {
		t.Fatal("expected mismatch")
	}
	DelDatabase(owner)
	if p.refs != 0 || !p.closed {
		t.Fatalf("failed registration leaked a pool reference: refs=%d closed=%v", p.refs, p.closed)
	}
}

func TestConcurrentRejectedBindingsDoNotLeakReferences(t *testing.T) {
	const owner = "concurrent-rejected-binding"
	p := &reviewPool{refs: 1}
	name2Pool.Store(owner, p)
	defer name2Pool.Delete(owner)
	defer name2LastPoolStats.Delete(owner)
	var workers sync.WaitGroup
	for range 32 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			if err := AddDatabase("unused", "unused", "private", "unused", "unused", 2, owner); err == nil {
				t.Error("mismatch accepted")
			}
		}()
	}
	workers.Wait()
	DelDatabase(owner)
	if p.refs != 0 || !p.closed {
		t.Fatal("rejected concurrent registrations retained references", p.refs, p.closed)
	}
}
