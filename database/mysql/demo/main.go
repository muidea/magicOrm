package main

import (
	"flag"
	log "github.com/cihub/seelog"
	"math/rand"
	"sync"
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
var itemSize = 10000000
var mode = 0
var finishFlag = false

type funcPtr func(executor *mysql.Executor) error

func main() {
	flag.StringVar(&databaseServer, "Server", databaseServer, "database server address")
	flag.StringVar(&databaseName, "Database", databaseName, "database name")
	flag.StringVar(&databaseUsername, "Account", databaseUsername, "database account")
	flag.StringVar(&databasePassword, "Password", databasePassword, "database password")
	flag.IntVar(&threadSize, "ThreadSize", threadSize, "database access thread size")
	flag.IntVar(&itemSize, "ItemSize", itemSize, "database insert item size")
	flag.IntVar(&mode, "Mode", mode, "database access mode, 1=onceDML,2=onceDDL,3=monkeyDDL")
	flag.Parse()
	if mode == 0 {
		flag.PrintDefaults()
		return
	}

	pool := mysql.NewPool()
	config := mysql.NewConfig(databaseServer, databaseName, databaseUsername, databasePassword)
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

		log.Infof("execute ok, elapse:%v", elapse)
	}()

	wg := &sync.WaitGroup{}

	switch mode {
	case onceDML:
		testDML(wg, pool)
	case onceDDL:
		testDDL(wg, pool)
	case monkeyDDL:
		testMonkey(wg, pool)
	}

	wg.Wait()
}

func testDML(wg *sync.WaitGroup, pool *mysql.Pool) {
	finishFlag = false
	wg.Add(1)
	pickExecutor(pool, wg, dropSchema)

	wg.Add(1)
	pickExecutor(pool, wg, createSchema)

	for idx := 0; idx < threadSize; idx++ {
		wg.Add(1)
		go pickExecutor(pool, wg, insertValue)
	}
	finishFlag = true
}

func testDDL(wg *sync.WaitGroup, pool *mysql.Pool) {
	wg.Add(1)
	pickExecutor(pool, wg, alterSchemaAdd)

	wg.Add(1)
	pickExecutor(pool, wg, alterSchemaOlnDDLAdd)

	wg.Add(1)
	pickExecutor(pool, wg, alterSchemaDrop)

	wg.Add(1)
	pickExecutor(pool, wg, alterSchemaOlnDDLDrop)
}

func testMonkey(wg *sync.WaitGroup, pool *mysql.Pool) {
	testDML(wg, pool)

	go func() {
		for !finishFlag {
			time.Sleep(time.Duration(rand.Int()%10) + time.Second*20)
			randomDDL(wg, pool)
		}
	}()
}

func randomDDL(wg *sync.WaitGroup, pool *mysql.Pool) {
	testDDL(wg, pool)
	wg.Add(1)
	pickExecutor(pool, wg, createSchemaDDL)

	wg.Add(1)
	pickExecutor(pool, wg, dropSchemaDDL)
}

func pickExecutor(pool *mysql.Pool, wg *sync.WaitGroup, fPtr funcPtr) {
	executorPtr, executorErr := pool.FetchOut()
	if executorErr != nil {
		return
	}

	if fPtr != nil {
		err := fPtr(executorPtr)
		if err != nil {
			panic(err)
		}
	}

	pool.PutIn(executorPtr)

	wg.Done()
}

func createSchema(executor *mysql.Executor) (err error) {
	sql := "CREATE TABLE `Unit` (\n\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`i8` TINYINT NOT NULL ,\n\t`i16` SMALLINT NOT NULL ,\n\t`i32` INT NOT NULL ,\n\t`i64` BIGINT NOT NULL ,\n\t`name` TEXT NOT NULL ,\n\t`value` FLOAT NOT NULL ,\n\t`f64` DOUBLE NOT NULL ,\n\t`ts` DATETIME NOT NULL ,\n\t`flag` TINYINT NOT NULL ,\n\t`iArray` TEXT NOT NULL ,\n\t`fArray` TEXT NOT NULL ,\n\t`strArray` TEXT NOT NULL ,\n\tPRIMARY KEY (`id`)\n)"
	_, _, err = executor.Execute(sql)
	return
}

func dropSchema(executor *mysql.Executor) (err error) {
	sql := "DROP TABLE `Unit`"
	_, _, err = executor.Execute(sql)
	return
}

func checkSchema(executor *mysql.Executor) (ret bool, err error) {
	ret, err = executor.CheckTableExist(`Unit`)
	return
}

func insertValue(executor *mysql.Executor) (err error) {
	sql := "INSERT INTO `Unit` (`i8`,`i16`,`i32`,`i64`,`name`,`value`,`f64`,`ts`,`flag`,`iArray`,`fArray`,`strArray`) VALUES (8,1600,323200,78962222222,'Hello world',12.345600128173828,12.45678,'2018-01-02 15:04:05',1,'12','12.34','abcdef')"
	idx := 0
	for idx < itemSize/threadSize {
		_, _, err = executor.Execute(sql)
		idx++
	}
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
	sql := "CREATE TABLE `Unit002` (\n\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`i8` TINYINT NOT NULL ,\n\t`i16` SMALLINT NOT NULL ,\n\t`i32` INT NOT NULL ,\n\t`i64` BIGINT NOT NULL ,\n\t`name` TEXT NOT NULL ,\n\t`value` FLOAT NOT NULL ,\n\t`f64` DOUBLE NOT NULL ,\n\t`ts` DATETIME NOT NULL ,\n\t`flag` TINYINT NOT NULL ,\n\t`iArray` TEXT NOT NULL ,\n\t`fArray` TEXT NOT NULL ,\n\t`strArray` TEXT NOT NULL ,\n\tPRIMARY KEY (`id`)\n), ALGORITHM=DEFAULT, LOCK=NONE"
	_, _, err = executor.Execute(sql)
	return
}

func dropSchemaDDL(executor *mysql.Executor) (err error) {
	sql := "DROP TABLE `Unit002`, ALGORITHM=DEFAULT, LOCK=NONE"
	_, _, err = executor.Execute(sql)
	return
}
