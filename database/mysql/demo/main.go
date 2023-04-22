package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/muidea/magicOrm/database/mysql"
)

const (
	onceDML   = 1
	onceDDL   = 2
	monkeyDDL = 3
)

var databaseServer = "localhost:3306"
var databaseName = "testdb"
var databaseUsername = "root"
var databasePassword = "rootkit"
var threadSize = 20

var itemSize = 18000000
var mode = 3
var disableStatic = true
var finishFlag = false

type funcPtr func(executor *mysql.Executor) error

var currentSize atomic.Int64

const truncateSize = 10000000

func main() {
	flag.StringVar(&databaseServer, "Server", databaseServer, "database server address")
	flag.StringVar(&databaseName, "Database", databaseName, "database name")
	flag.StringVar(&databaseUsername, "Account", databaseUsername, "database account")
	flag.StringVar(&databasePassword, "Password", databasePassword, "database password")
	flag.IntVar(&threadSize, "ThreadSize", threadSize, "database access thread size")
	flag.IntVar(&itemSize, "ItemSize", itemSize, "database insert item size")
	flag.IntVar(&mode, "Mode", mode, "database access mode, 1=onceDML,2=onceDDL,3=monkeyDDL")
	flag.BoolVar(&disableStatic, "DisableStatic", disableStatic, "disable static ddl,only online ddl")
	flag.Parse()
	if mode == 0 {
		flag.PrintDefaults()
		return
	}

	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		elapse := endTime.Sub(startTime)
		if err := recover(); err != nil {
			fmt.Printf("execute terminated, elapse:%v, err:%v", elapse, err)
			return
		}

		fmt.Printf("execute finished, elapse:%v", elapse)
	}()

	pool := mysql.NewPool()
	config := mysql.NewConfig(databaseServer, databaseName, databaseUsername, databasePassword)
	pool.Initialize(50, config)
	defer pool.Uninitialized()

	wg := &sync.WaitGroup{}
	finishFlag = false

	switch mode {
	case onceDML:
		testDML(wg, pool)
	case onceDDL:
		testDDL(wg, pool)
	case monkeyDDL:
		testMonkey(wg, pool)
	default:
	}

	wg.Wait()
	finishFlag = true
}

func testDML(wg *sync.WaitGroup, pool *mysql.Pool) {
	pickExecutor(pool, wg, dropSchema)

	pickExecutor(pool, wg, createSchema)

	for idx := 0; idx < threadSize; idx++ {
		go pickExecutor(pool, wg, insertValue)
	}
}

func testDDL(wg *sync.WaitGroup, pool *mysql.Pool) {
	if disableStatic {
		pickExecutor(pool, wg, alterSchemaAdd)
	}

	pickExecutor(pool, wg, alterSchemaOlnDDLAdd)

	if disableStatic {
		pickExecutor(pool, wg, alterSchemaDrop)
	}

	pickExecutor(pool, wg, alterSchemaOlnDDLDrop)
}

func testMonkey(wg *sync.WaitGroup, pool *mysql.Pool) {
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

func randomDDL(wg *sync.WaitGroup, pool *mysql.Pool) {
	testDDL(wg, pool)

	if currentSize.Load() > truncateSize {
		pickExecutor(pool, wg, truncateSchema)
	}

	pickExecutor(pool, wg, specialCheck)
	pickExecutor(pool, wg, createSchemaDDL)
	pickExecutor(pool, wg, dropSchemaDDL)
}

func pickExecutor(pool *mysql.Pool, wg *sync.WaitGroup, fPtr funcPtr) {
	var err error
	func() {
		wg.Add(1)
		defer wg.Done()

		executorPtr, executorErr := pool.FetchOut()
		if executorErr != nil {
			return
		}

		if fPtr != nil {
			err = fPtr(executorPtr)
		}

		pool.PutIn(executorPtr)
	}()

	if err != nil {
		panic(err)
	}
}

func createSchema(executor *mysql.Executor) (err error) {
	sql := "CREATE TABLE `Unit` (\n\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`i8` TINYINT NOT NULL ,\n\t`i16` SMALLINT NOT NULL ,\n\t`i32` INT NOT NULL ,\n\t`i64` BIGINT NOT NULL ,\n\t`name` TEXT NOT NULL ,\n\t`value` FLOAT NOT NULL ,\n\t`f64` DOUBLE NOT NULL ,\n\t`ts` DATETIME NOT NULL ,\n\t`flag` TINYINT NOT NULL ,\n\t`iArray` TEXT NOT NULL ,\n\t`fArray` TEXT NOT NULL ,\n\t`strArray` TEXT NOT NULL ,\n\tPRIMARY KEY (`id`)\n)"
	_, _, err = executor.Execute(sql)
	return
}

func dropSchema(executor *mysql.Executor) (err error) {
	sql := "DROP TABLE IF EXISTS `Unit`"
	_, _, err = executor.Execute(sql)
	return
}

func checkSchema(tableName string, executor *mysql.Executor) (ret bool, err error) {
	ret, err = executor.CheckTableExist(tableName)
	return
}

func insertValue(executor *mysql.Executor) (err error) {
	sql := "INSERT INTO `Unit` (`i8`,`i16`,`i32`,`i64`,`name`,`value`,`f64`,`ts`,`flag`,`iArray`,`fArray`,`strArray`) VALUES (8,1600,323200,78962222222,'Hello world',12.345600128173828,12.45678,'2018-01-02 15:04:05',1,'12','12.34','abcdef')"
	idx := 0
	if itemSize == -1 {
		for {
			_, _, err = executor.Execute(sql)
			if err == nil {
				currentSize.Add(1)
			}
		}

		return
	}

	for idx < itemSize/threadSize {
		_, _, err = executor.Execute(sql)
		if err == nil {
			currentSize.Add(1)
		}
		idx++
	}
	return
}

func truncateSchema(executor *mysql.Executor) (err error) {
	sql := "TRUNCATE TABLE `Unit`"
	_, _, err = executor.Execute(sql)

	currentSize.Swap(0)

	return
}

func alterSchemaAdd(executor *mysql.Executor) (err error) {
	sql := "ALTER TABLE `Unit` ADD dVal DATE"
	_, _, err = executor.Execute(sql)
	return
}

func alterSchemaDrop(executor *mysql.Executor) (err error) {
	sql := "ALTER TABLE `Unit` DROP dVal"
	_, _, err = executor.Execute(sql)
	return
}

func alterSchemaOlnDDLAdd(executor *mysql.Executor) (err error) {
	sql := "ALTER TABLE `Unit` ADD dVal2 DATE, ALGORITHM=DEFAULT, LOCK=NONE"
	_, _, err = executor.Execute(sql)
	return
}

func alterSchemaOlnDDLDrop(executor *mysql.Executor) (err error) {
	sql := "ALTER TABLE `Unit` DROP dVal2, ALGORITHM=DEFAULT, LOCK=NONE"
	_, _, err = executor.Execute(sql)
	return
}

func createSchemaDDL(executor *mysql.Executor) (err error) {
	sql := "CREATE TABLE `Unit002` (\n\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`i8` TINYINT NOT NULL ,\n\t`i16` SMALLINT NOT NULL ,\n\t`i32` INT NOT NULL ,\n\t`i64` BIGINT NOT NULL ,\n\t`name` TEXT NOT NULL ,\n\t`value` FLOAT NOT NULL ,\n\t`f64` DOUBLE NOT NULL ,\n\t`ts` DATETIME NOT NULL ,\n\t`flag` TINYINT NOT NULL ,\n\t`iArray` TEXT NOT NULL ,\n\t`fArray` TEXT NOT NULL ,\n\t`strArray` TEXT NOT NULL ,\n\tPRIMARY KEY (`id`)\n)"
	_, _, err = executor.Execute(sql)
	return
}

func dropSchemaDDL(executor *mysql.Executor) (err error) {
	sql := "DROP TABLE IF EXISTS `Unit002`"
	_, _, err = executor.Execute(sql)
	return
}

func specialCheck(executor *mysql.Executor) (err error) {
	ok, _ := checkSchema("rbac_menuinfo", executor)
	if ok {
		alterSpecialSchemaOlnDDLAdd(executor)
		alterSpecialSchemaOlnDDLDrop(executor)
	}

	return
}

func alterSpecialSchemaOlnDDLAdd(executor *mysql.Executor) (err error) {
	sql := "ALTER TABLE `rbac_menuinfo` ADD dVal200 DATE, ALGORITHM=DEFAULT, LOCK=NONE"
	_, _, err = executor.Execute(sql)
	return
}

func alterSpecialSchemaOlnDDLDrop(executor *mysql.Executor) (err error) {
	sql := "ALTER TABLE `rbac_menuinfo` DROP dVal200, ALGORITHM=DEFAULT, LOCK=NONE"
	_, _, err = executor.Execute(sql)
	return
}
