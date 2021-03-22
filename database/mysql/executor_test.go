package mysql

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	pool := NewPool()

	pool.Initialize(50, "root", "rootkit", "localhost:3306", "testdb")
	defer pool.Uninitialize()

	wg := &sync.WaitGroup{}

	for idx := 0; idx < 2000; idx++ {
		wg.Add(1)
		go pickExecutor(pool, wg)
	}

	wg.Wait()
}

func pickExecutor(pool *Pool, wg *sync.WaitGroup) {
	executorPtr, executorErr := pool.FetchOut()
	if executorErr != nil {
		return
	}

	time.Sleep(time.Duration(rand.Int()%10) + time.Second)

	pool.PutIn(executorPtr)

	wg.Done()
}
