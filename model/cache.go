package model

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
	kvCache cache.KVCache
}

// NewCache new modelInfo cache
func NewCache() Cache {
	return &impl{kvCache: cache.NewKVCache(nil)}
}

func (s *impl) Reset() {
	s.kvCache.ClearAll()
}

func (s *impl) Put(name string, vModel Model) {
	s.kvCache.Put(name, vModel, cache.MaxAgeValue)
}

func (s *impl) Fetch(name string) (ret Model) {
	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	val := s.kvCache.Fetch(name)
	if val == nil {
		return nil
	}

	ret = val.(Model)
	return
}

func (s *impl) Remove(name string) {
	s.kvCache.Remove(name)
}
