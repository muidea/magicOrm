package cache

import (
	"errors"
	"sync"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	verrors "github.com/muidea/magicOrm/validation/errors"
)

type cacheDirective struct {
	key  models.Key
	args []string
}

func (d cacheDirective) Key() models.Key { return d.key }
func (d cacheDirective) Args() []string  { return d.args }
func (d cacheDirective) HasArgs() bool   { return len(d.args) > 0 }

type cacheConstraints struct {
	directives []models.Directive
}

func (c cacheConstraints) Has(key models.Key) bool {
	_, ok := c.Get(key)
	return ok
}
func (c cacheConstraints) Get(key models.Key) (models.Directive, bool) {
	for _, directive := range c.directives {
		if directive.Key() == key {
			return directive, true
		}
	}
	return nil, false
}
func (c cacheConstraints) Directives() []models.Directive { return c.directives }

type cacheModel struct{ name string }

func (m *cacheModel) GetName() string                      { return m.name }
func (m *cacheModel) GetShowName() string                  { return m.name }
func (m *cacheModel) GetPkgPath() string                   { return "validation.cache" }
func (m *cacheModel) GetPkgKey() string                    { return m.GetPkgPath() + "/" + m.name }
func (m *cacheModel) GetDescription() string               { return m.name }
func (m *cacheModel) GetFields() models.Fields             { return nil }
func (m *cacheModel) SetFieldValue(string, any) *cd.Error  { return nil }
func (m *cacheModel) SetPrimaryFieldValue(any) *cd.Error   { return nil }
func (m *cacheModel) GetPrimaryField() models.Field        { return nil }
func (m *cacheModel) GetField(string) models.Field         { return nil }
func (m *cacheModel) Interface(bool) any                   { return nil }
func (m *cacheModel) Copy(models.ViewDeclare) models.Model { return m }
func (m *cacheModel) Reset()                               {}

func TestConstraintCacheLifecycle(t *testing.T) {
	cache := NewConstraintCache(2, time.Millisecond*20)
	constraints := cacheConstraints{directives: []models.Directive{
		cacheDirective{key: models.KeyRequired},
		cacheDirective{key: models.KeyMin, args: []string{"3"}},
	}}

	key := cache.GenerateCacheKey("abc", constraints, verrors.ScenarioInsert)
	cache.Set(key, "abc", constraints, verrors.ScenarioInsert, nil)

	if result, ok := cache.Get(key); !ok || result != nil {
		t.Fatalf("expected cached successful result, got ok=%v result=%v", ok, result)
	}

	stats := cache.GetStats()
	if stats.Hits != 1 || stats.Size != 1 {
		t.Fatalf("unexpected stats after hit: %+v", stats)
	}

	cache.Set(cache.GenerateCacheKey("first", constraints, verrors.ScenarioInsert), "first", constraints, verrors.ScenarioInsert, errors.New("first"))
	cache.Set(cache.GenerateCacheKey("second", constraints, verrors.ScenarioInsert), "second", constraints, verrors.ScenarioInsert, errors.New("second"))
	if cache.GetStats().Evictions == 0 {
		t.Fatal("expected LRU eviction when cache exceeds max size")
	}

	time.Sleep(time.Millisecond * 25)
	if _, ok := cache.Get(key); ok {
		t.Fatal("expected expired cache entry to miss")
	}

	cache.ClearExpired()
	if cache.GetStats().Size != 0 {
		t.Fatalf("expected expired entries to be cleared, stats=%+v", cache.GetStats())
	}
}

func TestValidationCacheLifecycle(t *testing.T) {
	cfg := DefaultCacheConfig()
	cfg.DefaultTTL = time.Millisecond * 20
	cfg.CleanupInterval = 0

	validationCache := NewValidationCache(cfg)
	constraints := cacheConstraints{directives: []models.Directive{
		cacheDirective{key: models.KeyRequired},
	}}

	validationCache.SetConstraintResult("value", constraints, verrors.ScenarioInsert, nil)
	if _, ok := validationCache.GetConstraintResult("value", constraints, verrors.ScenarioInsert); !ok {
		t.Fatal("expected constraint cache hit")
	}

	model := &cacheModel{name: "User"}
	modelErr := errors.New("model invalid")
	validationCache.SetModelResult(model, verrors.ScenarioUpdate, modelErr)
	if result, ok := validationCache.GetModelResult(model, verrors.ScenarioUpdate); !ok || result != modelErr {
		t.Fatalf("expected model cache hit, got ok=%v result=%v", ok, result)
	}

	validationCache.SetModelResult(nil, verrors.ScenarioDelete, nil)
	if _, ok := validationCache.GetModelResult(nil, verrors.ScenarioDelete); !ok {
		t.Fatal("expected nil model cache key to be supported")
	}

	stats := validationCache.GetStats()
	if !stats["enabled"].(bool) {
		t.Fatal("expected cache stats to report enabled")
	}

	validationCache.Disable()
	if validationCache.IsEnabled() {
		t.Fatal("expected cache to be disabled")
	}
	validationCache.Enable()
	if !validationCache.IsEnabled() {
		t.Fatal("expected cache to be re-enabled")
	}

	time.Sleep(time.Millisecond * 25)
	validationCache.ClearExpired()
	if _, ok := validationCache.GetModelResult(model, verrors.ScenarioUpdate); ok {
		t.Fatal("expected model cache entry to expire")
	}

	validationCache.Clear()
	modelStats := validationCache.GetStats()["model_cache"].(map[string]interface{})
	if modelStats["size"].(int) != 0 {
		t.Fatalf("expected cleared model cache, got %+v", modelStats)
	}
}

func TestConstraintCacheSeparatesValuesTypesAndArgumentBoundaries(t *testing.T) {
	c := NewConstraintCache(100, time.Minute)
	rule := cacheConstraints{directives: []models.Directive{cacheDirective{key: models.KeyMin, args: []string{"1"}}}}
	pairs := [][2]any{{int32(5), int32(0)}, {int(1), int64(1)}, {true, false}, {float64(1), float64(2)}, {[]byte("a"), []byte("b")}, {[]byte(nil), []byte{}}, {nil, int(0)}}
	for _, p := range pairs {
		a, b := c.GenerateCacheKey(p[0], rule, verrors.ScenarioInsert), c.GenerateCacheKey(p[1], rule, verrors.ScenarioInsert)
		if a == "" || b == "" || a == b {
			t.Fatalf("different inputs shared or lost cache key: %T/%T", p[0], p[1])
		}
		c.Set(a, p[0], rule, verrors.ScenarioInsert, nil)
		if _, ok := c.Get(b); ok {
			t.Fatal("validation success leaked to a different value")
		}
	}
	one := cacheConstraints{directives: []models.Directive{cacheDirective{key: models.KeyIn, args: []string{"a:b"}}}}
	two := cacheConstraints{directives: []models.Directive{cacheDirective{key: models.KeyIn, args: []string{"a", "b"}}}}
	if c.GenerateCacheKey("x", one, verrors.ScenarioInsert) == c.GenerateCacheKey("x", two, verrors.ScenarioInsert) {
		t.Fatal("directive argument boundaries collided")
	}
	if c.GenerateCacheKey(1, rule, verrors.ScenarioInsert) == c.GenerateCacheKey(1, rule, verrors.ScenarioUpdate) {
		t.Fatal("scenario keys collided")
	}
}

func TestConstraintCacheBypassesComplexValues(t *testing.T) {
	c := NewConstraintCache(10, time.Minute)
	value := 1
	for _, v := range []any{&value, (*int)(nil), []int{1}, map[string]int{"x": 1}, struct{ X int }{1}} {
		k := c.GenerateCacheKey(v, nil, verrors.ScenarioInsert)
		if k != "" {
			t.Fatalf("unsafe complex value cached: %T", v)
		}
		c.Set(k, v, nil, verrors.ScenarioInsert, nil)
		if _, ok := c.Get(k); ok {
			t.Fatal("bypassed value produced a cache hit")
		}
	}
	if c.GetStats().Size != 0 {
		t.Fatal("bypassed values retained")
	}
}

func TestConstraintCacheConcurrentHits(t *testing.T) {
	c := NewConstraintCache(10, time.Minute)
	key := c.GenerateCacheKey(1, nil, verrors.ScenarioInsert)
	c.Set(key, 1, nil, verrors.ScenarioInsert, nil)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				c.Get(key)
				c.GetStats()
			}
		}()
	}
	wg.Wait()
	if c.GetStats().Hits != 800 {
		t.Fatal("lost concurrent cache hits")
	}
}

func TestModelCacheEvictionAndCleanupTicker(t *testing.T) {
	modelCache := &ModelCache{
		cache: map[string]*ModelCacheEntry{
			"oldest": {Timestamp: time.Now().Add(-time.Minute)},
			"newest": {Timestamp: time.Now()},
		},
		maxSize:    2,
		defaultTTL: time.Millisecond,
	}
	modelCache.evictOldest()
	if _, exists := modelCache.cache["oldest"]; exists {
		t.Fatal("expected oldest model cache entry to be evicted")
	}

	cfg := DefaultCacheConfig()
	cfg.DefaultTTL = time.Millisecond
	cfg.CleanupInterval = time.Millisecond
	validationCache := NewValidationCache(cfg)
	validationCache.SetModelResult(&cacheModel{name: "Ticker"}, verrors.ScenarioInsert, nil)
	time.Sleep(time.Millisecond * 5)
	if _, ok := validationCache.GetModelResult(&cacheModel{name: "Ticker"}, verrors.ScenarioInsert); ok {
		t.Fatal("expected cleanup ticker to remove expired model entry")
	}
}
