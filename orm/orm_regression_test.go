package orm

import "testing"

func TestDefaultValidationConfigDisablesCachingForTransientORM(t *testing.T) {
	cfg := defaultValidationConfig(false)
	if cfg.EnableCaching {
		t.Fatal("expected transient ORM validation caching to be disabled")
	}
}
