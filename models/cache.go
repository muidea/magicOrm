package models

import (
	"github.com/muidea/magicCommon/foundation/cache"
)

// Cache Model Cache
type Cache interface {
	Reset()

	Put(name string, vModel Model)

	Fetch(name string) Model

	Remove(name string)
}

type impl struct {
	kvCache cache.KVCacheGeneric[string, Model]
}

// NewCache new modelInfo cache
func NewCache() Cache {
	return &impl{kvCache: cache.NewGenericKVCache[string, Model](nil)}
}

func (s *impl) Reset() {
	s.kvCache.ClearAll()
}

func (s *impl) Put(name string, vModel Model) {
	s.kvCache.Put(name, vModel, cache.ForeverAgeValue)
}

func (s *impl) Fetch(name string) (ret Model) {
	ret = s.kvCache.Fetch(name)
	return
}

func (s *impl) Remove(name string) {
	s.kvCache.Remove(name)
}
