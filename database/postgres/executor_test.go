package postgres

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
)

const (
	onceDML   = 1
	onceDDL   = 2
	monkeyDDL = 3
)

var databaseServer = "localhost:5432"
var databaseName = "magicplatform_db"
var databaseUsername = "postgres"
var databasePassword = "rootkit"
var threadSize = 20
var itemSize = 1
var mode = 1

var finishFlag = false

type funcPtr func(executor *ConnExecutor) *cd.Error

func TestNewPool(t *testing.T) {
	pool := NewPool()
	config := NewConfig(databaseServer, databaseName, databaseUsername, databasePassword, "")
	pool.Initialize(50, config)
	defer pool.Uninitialized()

	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		elapse := endTime.Sub(startTime)
		if err := recover(); err != nil {
			log.Errorf("execute failed, elapse:%v, err:%v", elapse, err)
			return
		}
	}()

	wg := &sync.WaitGroup{}

	finishFlag = false
	switch mode {
	case onceDML:
		testDML(wg, pool)
	case onceDDL:
		testDDL(wg, pool)
	case monkeyDDL:
		testMonkey(wg, pool)
	}

	wg.Wait()
	finishFlag = true
}

func testDML(wg *sync.WaitGroup, pool *Pool) {
	wg.Add(1)
	pickExecutor(pool, wg, dropSchema)

	wg.Add(1)
	pickExecutor(pool, wg, createSchema)

	wg.Add(1)
	pickExecutor(pool, wg, func(executor *ConnExecutor) (err *cd.Error) {
		bVal, bErr := checkSchema(executor)
		if bErr != nil {
			log.Errorf("checkSchema failed, error:%s", bErr.Error())
			err = bErr
			return
		}
		if !bVal {
			log.Errorf("checkSchema failed")
		}

		return
	})

	for idx := 0; idx < threadSize; idx++ {
		wg.Add(1)
		go pickExecutor(pool, wg, insertValue)
	}
}

func testDDL(wg *sync.WaitGroup, pool *Pool) {
	wg.Add(1)
	pickExecutor(pool, wg, alterSchemaAdd)

	wg.Add(1)
	pickExecutor(pool, wg, alterSchemaOlnDDLAdd)

	wg.Add(1)
	pickExecutor(pool, wg, alterSchemaDrop)

	wg.Add(1)
	pickExecutor(pool, wg, alterSchemaOlnDDLDrop)
}

func testMonkey(wg *sync.WaitGroup, pool *Pool) {
	testDML(wg, pool)

	go func() {
		for !finishFlag {
			time.Sleep(time.Duration(rand.Int()%10) + time.Second*20)
			if finishFlag {
				break
			}

			randomDDL(wg, pool)
		}
	}()
}

func randomDDL(wg *sync.WaitGroup, pool *Pool) {
	testDDL(wg, pool)
	wg.Add(1)
	pickExecutor(pool, wg, createSchemaDDL)

	wg.Add(1)
	pickExecutor(pool, wg, dropSchemaDDL)
}

func pickExecutor(pool *Pool, wg *sync.WaitGroup, fPtr funcPtr) {
	executorPtr, executorErr := pool.GetExecutor(context.Background())
	defer executorPtr.Release()

	if executorErr != nil {
		return
	}

	if fPtr != nil {
		err := fPtr(executorPtr)
		if err != nil {
			panic(err)
		}
	}

	wg.Done()
}

func createSchema(executor *ConnExecutor) (err *cd.Error) {
	sql := "CREATE TABLE \"Unit001\" (\n\t\"id\" SERIAL PRIMARY KEY,\n\t\"i8\" SMALLINT NOT NULL ,\n\t\"i16\" SMALLINT NOT NULL ,\n\t\"i32\" INTEGER NOT NULL ,\n\t\"i64\" BIGINT NOT NULL ,\n\t\"name\" TEXT NOT NULL ,\n\t\"value\" REAL NOT NULL ,\n\t\"f64\" DOUBLE PRECISION NOT NULL ,\n\t\"ts\" TIMESTAMP NOT NULL ,\n\t\"flag\" SMALLINT NOT NULL ,\n\t\"iArray\" TEXT NOT NULL ,\n\t\"fArray\" TEXT NOT NULL ,\n\t\"strArray\" TEXT NOT NULL \n)"
	_, err = executor.Execute(sql)
	return
}

func dropSchema(executor *ConnExecutor) (err *cd.Error) {
	sql := "DROP TABLE IF EXISTS \"Unit001\""
	_, err = executor.Execute(sql)
	return
}

func checkSchema(executor *ConnExecutor) (ret bool, err *cd.Error) {
	ret, err = executor.CheckTableExist("Unit001")
	return
}

func insertValue(executor *ConnExecutor) (err *cd.Error) {
	sql := "INSERT INTO \"Unit001\" (\"i8\",\"i16\",\"i32\",\"i64\",\"name\",\"value\",\"f64\",\"ts\",\"flag\",\"iArray\",\"fArray\",\"strArray\") VALUES (8,1600,323200,78962222222,'Hello world',12.345600128173828,12.45678,'2018-01-02 15:04:05',1,'12','12.34','abcdef')"
	idx := 0
	for idx < itemSize/threadSize {
		_, err = executor.Execute(sql)
		idx++
	}
	return
}

func alterSchemaAdd(executor *ConnExecutor) (err *cd.Error) {
	sql := "ALTER TABLE \"Unit001\" ADD dVal DATE"
	_, err = executor.Execute(sql)
	return
}

func alterSchemaDrop(executor *ConnExecutor) (err *cd.Error) {
	sql := "ALTER TABLE \"Unit001\" DROP dVal"
	_, err = executor.Execute(sql)
	return
}

func alterSchemaOlnDDLAdd(executor *ConnExecutor) (err *cd.Error) {
	sql := "ALTER TABLE \"Unit001\" ADD dVal2 DATE"
	_, err = executor.Execute(sql)
	return
}

func alterSchemaOlnDDLDrop(executor *ConnExecutor) (err *cd.Error) {
	sql := "ALTER TABLE \"Unit001\" DROP dVal2"
	_, err = executor.Execute(sql)
	return
}

func createSchemaDDL(executor *ConnExecutor) (err *cd.Error) {
	sql := "CREATE TABLE \"Unit002\" (\n\t\"id\" SERIAL PRIMARY KEY,\n\t\"i8\" SMALLINT NOT NULL ,\n\t\"i16\" SMALLINT NOT NULL ,\n\t\"i32\" INTEGER NOT NULL ,\n\t\"i64\" BIGINT NOT NULL ,\n\t\"name\" TEXT NOT NULL ,\n\t\"value\" REAL NOT NULL ,\n\t\"f64\" DOUBLE PRECISION NOT NULL ,\n\t\"ts\" TIMESTAMP NOT NULL ,\n\t\"flag\" SMALLINT NOT NULL ,\n\t\"iArray\" TEXT NOT NULL ,\n\t\"fArray\" TEXT NOT NULL ,\n\t\"strArray\" TEXT NOT NULL \n)"
	_, err = executor.Execute(sql)
	return
}

func dropSchemaDDL(executor *ConnExecutor) (err *cd.Error) {
	sql := "DROP TABLE IF EXISTS \"Unit002\""
	_, err = executor.Execute(sql)
	return
}
