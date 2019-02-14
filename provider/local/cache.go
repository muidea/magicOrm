package local

import (
	"muidea.com/magicCommon/foundation/cache"
	"muidea.com/magicOrm/model"
)

// Cache Model Cache
type Cache interface {
	Reset()

	Put(name string, info model.Model)

	Fetch(name string) model.Model

	Remove(name string)
}

type impl struct {
	kvCache cache.KVCache
}

// NewCache new modelInfo cache
func NewCache() Cache {
	return &impl{kvCache: cache.NewKVCache()}
}

func (s *impl) Reset() {
	s.kvCache.ClearAll()
}

func (s *impl) Put(name string, info model.Model) {
	s.kvCache.Put(name, info, cache.MaxAgeValue)
}

func (s *impl) Fetch(name string) model.Model {
	obj, ok := s.kvCache.Fetch(name)
	if !ok {
		return nil
	}

	return obj.(model.Model)
}

func (s *impl) Remove(name string) {
	s.kvCache.Remove(name)
}
