package cache

import (
	"time"
)

// KVCache 缓存对象
type KVCache interface {
	// maxAge单位minute
	Put(key string, data interface{}, maxAge float64) string
	Fetch(key string) interface{}
	Search(opr SearchOpr) interface{}
	Remove(key string)
	GetAll() []interface{}
	ClearAll()
	Release()
}

// NewKVCache 创建Cache对象
func NewKVCache() KVCache {
	cache := make(MemoryKVCache)

	go cache.run()
	go cache.checkTimeOut()

	return &cache
}

type putInKVData struct {
	key    string
	data   interface{}
	maxAge float64
}

type putInKVResult struct {
	value string
}

type fetchOutKVData struct {
	key string
}

type fetchOutKVResult struct {
	value interface{}
}

type searchKVData struct {
	opr SearchOpr
}

type searchKVResult fetchOutKVResult

type getAllKVResult struct {
	value []interface{}
}

type removeKVData struct {
	key string
}

type cacheKVData struct {
	putInKVData
	cacheTime time.Time
}

// MemoryKVCache 内存缓存
type MemoryKVCache chan commandData

// Put 投放数据，返回数据的唯一标示
func (right *MemoryKVCache) Put(key string, data interface{}, maxAge float64) string {
	reply := make(chan interface{})

	putInData := &putInKVData{}
	putInData.key = key
	putInData.data = data
	putInData.maxAge = maxAge

	*right <- commandData{action: putIn, value: putInData, result: reply}

	result := (<-reply).(*putInKVResult).value
	return result
}

// Fetch 获取数据
func (right *MemoryKVCache) Fetch(key string) interface{} {
	reply := make(chan interface{})

	fetchOutData := &fetchOutKVData{}
	fetchOutData.key = key

	*right <- commandData{action: fetchOut, value: fetchOutData, result: reply}

	result := (<-reply).(*fetchOutKVResult)
	return result.value
}

func (right *MemoryKVCache) Search(opr SearchOpr) interface{} {
	if opr == nil {
		return nil
	}

	reply := make(chan interface{})

	searchData := &searchKVData{}
	searchData.opr = opr

	*right <- commandData{action: search, value: searchData, result: reply}

	result := (<-reply).(*searchKVResult)
	return result.value
}

// Remove 清除数据
func (right *MemoryKVCache) Remove(key string) {
	removeKVData := &removeKVData{}
	removeKVData.key = key

	*right <- commandData{action: remove, value: removeKVData}
}

// GetAll 获取所有的数据
func (right *MemoryKVCache) GetAll() (ret []interface{}) {
	reply := make(chan interface{})

	*right <- commandData{action: getAll, value: nil, result: reply}

	result := (<-reply).(*getAllKVResult)

	ret = result.value

	return
}

// ClearAll 清除所有数据
func (right *MemoryKVCache) ClearAll() {

	*right <- commandData{action: clearAll}
}

// Release 释放Cache
func (right *MemoryKVCache) Release() {
	*right <- commandData{action: end}

	close(*right)
}

func (right *MemoryKVCache) run() {
	localCacheData := make(map[string]cacheKVData)

	for command := range *right {
		switch command.action {
		case putIn:
			cacheKVData := cacheKVData{}
			cacheKVData.putInKVData = *(command.value.(*putInKVData))
			cacheKVData.cacheTime = time.Now()

			localCacheData[cacheKVData.key] = cacheKVData

			result := &putInKVResult{}
			result.value = cacheKVData.key

			command.result <- result
		case fetchOut:
			key := command.value.(*fetchOutKVData).key

			cacheKVData, found := localCacheData[key]

			result := &fetchOutKVResult{}
			if found {
				cacheKVData.cacheTime = time.Now()
				localCacheData[key] = cacheKVData

				result.value = cacheKVData.data
			}

			command.result <- result
		case search:
			opr := command.value.(*searchKVData).opr

			result := &searchKVResult{}
			for _, v := range localCacheData {
				if opr(v.data) {
					result.value = v.data
					break
				}
			}

			command.result <- result
		case remove:
			key := command.value.(*removeKVData).key

			delete(localCacheData, key)
		case getAll:
			result := &getAllKVResult{value: []interface{}{}}
			for _, v := range localCacheData {
				result.value = append(result.value, v.data)
			}

			command.result <- result
		case clearAll:
			localCacheData = make(map[string]cacheKVData)
		case checkTimeOut:
			// 检查每项数据是否超时，超时数据需要主动清除掉
			for k, v := range localCacheData {
				if v.maxAge != MaxAgeValue {
					current := time.Now()
					elapse := current.Sub(v.cacheTime).Minutes()
					if elapse > v.maxAge {
						delete(localCacheData, k)
					}
				}
			}
		case end:
			localCacheData = nil
		}
	}
}

func (right *MemoryKVCache) checkTimeOut() {
	timeOutTimer := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timeOutTimer.C:
			*right <- commandData{action: checkTimeOut}
		}
	}
}