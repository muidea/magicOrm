package test

import "github.com/muidea/magicOrm/orm"

var config = orm.NewConfig("localhost:5432", "magicplatform_db", "postgres", "rootkit", "")
