package main

import (
	"flag"
	"sync"

	"github.com/muidea/magicOrm/database/mysql"
)

var databaseServer = "localhost:3306"
var databaseName = "testdb"
var databaseUsername = "root"
var databasePassword = "rootkit"
var threadSize = 20
var itemSize = 10000000
var mode = 1

type funcPtr func(executor *mysql.Executor) error

func main() {
	flag.StringVar(&databaseServer, "Server", databaseServer, "database server address")
	flag.StringVar(&databaseName, "Database", databaseName, "database name")
	flag.StringVar(&databaseUsername, "Account", databaseUsername, "database account")
	flag.StringVar(&databasePassword, "Password", databasePassword, "database password")
	flag.IntVar(&threadSize, "ThreadSize", threadSize, "database access thread size")
	flag.IntVar(&itemSize, "ItemSize", itemSize, "database insert item size")
	flag.IntVar(&mode, "Mode", mode, "database access mode")
	flag.Parse()

	pool := mysql.NewPool()
	config := mysql.NewConfig(databaseServer, databaseName, databaseUsername, databasePassword)
	pool.Initialize(50, config)
	defer pool.Uninitialized()

	wg := &sync.WaitGroup{}

	func() {
		wg.Add(1)
		pickExecutor(pool, wg, alterSchema)

		wg.Add(1)
		pickExecutor(pool, wg, alterSchemaOlnDDL)
	}()

	func() {
		wg.Add(1)
		pickExecutor(pool, wg, dropSchema)

		wg.Add(1)
		pickExecutor(pool, wg, createSchema)
	}()

	for idx := 0; idx < threadSize; idx++ {
		wg.Add(1)
		go pickExecutor(pool, wg, insertValue)
	}

	wg.Wait()
}

func pickExecutor(pool *mysql.Pool, wg *sync.WaitGroup, fPtr funcPtr) {
	executorPtr, executorErr := pool.FetchOut()
	if executorErr != nil {
		return
	}

	if fPtr != nil {
		fPtr(executorPtr)
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

func alterSchema(executor *mysql.Executor) (err error) {
	sql := "ALTER TABLE `Unit` ADD dVal DATE"
	_, _, err = executor.Execute(sql)
	return
}

func alterSchemaOlnDDL(executor *mysql.Executor) (err error) {
	sql := "ALTER TABLE `Unit` ADD dVal2 DATE, ALGORITHM=DEFAULT, LOCK=NONE"
	_, _, err = executor.Execute(sql)
	return
}
