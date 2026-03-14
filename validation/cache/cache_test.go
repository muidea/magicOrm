package cache

import (
	"errors"
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

func TestGetTypeHash(t *testing.T) {
	testCases := map[string]any{
		"string": "abc",
		"int":    1,
		"uint":   uint(1),
		"float":  1.2,
		"bool":   true,
		"bytes":  []byte("abc"),
		"other":  struct{}{},
	}

	expected := map[string]string{
		"string": "string:abc",
		"int":    "int",
		"uint":   "uint",
		"float":  "float",
		"bool":   "bool",
		"bytes":  "bytes",
		"other":  "complex",
	}

	for name, value := range testCases {
		if got := getTypeHash(value); got != expected[name] {
			t.Fatalf("%s: unexpected hash %q", name, got)
		}
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
