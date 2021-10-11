package mysql

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	pool := NewPool()
	config := NewConfig("localhost:3306", "testdb", "root", "rootkit")
	pool.Initialize(50, config)
	defer pool.Uninitialize()

	wg := &sync.WaitGroup{}

	for idx := 0; idx < 2000; idx++ {
		wg.Add(1)
		go pickExecutor(pool, wg)
	}

	wg.Wait()
}

func pickExecutor(pool *Pool, wg *sync.WaitGroup) {
	executorPtr, executorErr := pool.fetchOut()
	if executorErr != nil {
		return
	}

	time.Sleep(time.Duration(rand.Int()%10) + time.Second)

	pool.putIn(executorPtr)

	wg.Done()
}
