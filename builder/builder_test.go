package builder

import (
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/util"

	"github.com/muidea/magicOrm/provider"
)

// Unit 单元信息
type Unit struct {
	ID        string    `orm:"uid key uuid"`
	Name      string    `orm:"name"`
	Value     float64   `orm:"value"`
	TimeStamp time.Time `orm:"ts"`
}

type Ext struct {
	//ID 唯一标示单元
	ID int `orm:"eid key auto"`
	// Name 名称
	Name string `orm:"name"`

	Description *string `orm:"description"`

	Unit Unit `orm:"unit"`
}

func TestBuilderCommon(t *testing.T) {
	now, _ := time.ParseInLocation(util.CSTLayout, "2018-01-02 15:04:05", time.Local)
	unit := &Unit{ID: "10", Name: "Hello world", Value: 12.3456, TimeStamp: now}

	localProvider := provider.NewLocalProvider("default")
	_, err := localProvider.RegisterModel(unit)
	if err != nil {
		t.Errorf("localProvider.RegisterModel failed, err:%s", err.Error())
		return
	}

	info, err := localProvider.GetEntityModel(unit)
	if err != nil {
		t.Errorf("GetEntityModel failed, err:%s", err.Error())
		return
	}

	filter, err := localProvider.GetModelFilter(info)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	builder := NewBuilder(info, localProvider, "abc")
	if builder == nil {
		t.Error("new Builder failed")
		return
	}

	str, err := builder.BuildCreateTable()
	if err != nil {
		t.Errorf("build create schema failed, err:%s", err.Error())
		return
	}
	if str != "CREATE TABLE IF NOT EXISTS `abc_Unit` (\n\t`uid` TEXT NOT NULL ,\n\t`name` TEXT NOT NULL ,\n\t`value` DOUBLE NOT NULL ,\n\t`ts` DATETIME NOT NULL ,\n\tPRIMARY KEY (`uid`)\n)\n" {
		t.Errorf("build create schema failed, str:%s", str)
		return
	}

	str, err = builder.BuildDropTable()
	if err != nil {
		t.Errorf("build drop schema failed, err:%s", err.Error())
		return
	}
	if str != "DROP TABLE IF EXISTS `abc_Unit`" {
		t.Error("build drop schema failed")
		return
	}

	str, err = builder.BuildInsert()
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
	}
	if str != "INSERT INTO `abc_Unit` (`uid`,`name`,`value`,`ts`) VALUES ('10','Hello world',12.3456,'2018-01-02 15:04:05')" {
		t.Errorf("build insert failed, str:%s", str)
		return
	}

	str, err = builder.BuildUpdate()
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
		return
	}
	if str != "UPDATE `abc_Unit` SET `name`='Hello world',`value`=12.3456,`ts`='2018-01-02 15:04:05' WHERE `uid`='10'" {
		t.Errorf("build update failed, str:%s", str)
		return
	}

	str, err = builder.BuildDelete()
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
		return
	}
	if str != "DELETE FROM `abc_Unit` WHERE `uid`='10'" {
		t.Errorf("build delete failed, str:%s", str)
		return
	}

	err = filter.Above("value", 12)
	if err != nil {
		t.Errorf("filter.Above failed, err:%s", err.Error())
		return
	}
	str, err = builder.BuildQuery(filter)
	if err != nil {
		t.Errorf("build query failed, err:%s", err.Error())
		return
	}
	if str != "SELECT `uid`,`name`,`value`,`ts` FROM `abc_Unit` WHERE `value` > 12" {
		t.Errorf("build query failed, str:%s", str)
		return
	}

	str, err = builder.BuildCount(nil)
	if err != nil {
		t.Errorf("build count failed, err:%s", err.Error())
		return
	}
	if str != "SELECT COUNT(`uid`) FROM `abc_Unit`" {
		t.Errorf("build count failed, str:%s", str)
		return
	}

	str, err = builder.BuildCount(filter)
	if err != nil {
		t.Errorf("build count failed, err:%s", err.Error())
		return
	}
	if str != "SELECT COUNT(`uid`) FROM `abc_Unit` WHERE `value` > 12" {
		t.Errorf("build count failed, str:%s", str)
		return
	}

	return
}

func TestBuilderReference(t *testing.T) {
	var desc string
	ext := &Ext{ID: 12, Description: &desc}
	unit := &Unit{ID: "10"}

	localProvider := provider.NewLocalProvider("default")
	extModel, extErr := localProvider.RegisterModel(ext)
	if extErr != nil {
		t.Errorf("localProvider.RegisterModel failed, err:%s", extErr.Error())
		return
	}
	unitModel, unitErr := localProvider.RegisterModel(unit)
	if unitErr != nil {
		t.Errorf("localProvider.RegisterModel failed, err:%s", unitErr.Error())
	}

	builder := NewBuilder(extModel, localProvider, "abc")
	if builder == nil {
		t.Error("new Builder failed")
	}

	str, err := builder.BuildCreateTable()
	if err != nil {
		t.Errorf("build create schema failed, err:%s", err.Error())
	}
	if str != "CREATE TABLE IF NOT EXISTS `abc_Ext` (\n\t`eid` INT NOT NULL AUTO_INCREMENT,\n\t`name` TEXT NOT NULL ,\n\t`description` TEXT NOT NULL ,\n\tPRIMARY KEY (`eid`)\n)\n" {
		t.Errorf("build create schema failed, str:%s", str)
	}

	str, err = builder.BuildDropTable()
	if err != nil {
		t.Errorf("build drop schema failed, err:%s", err.Error())
	}
	if str != "DROP TABLE IF EXISTS `abc_Ext`" {
		t.Errorf("build drop schema failed, str:%s", str)
	}

	str, err = builder.BuildInsert()
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
	}
	if str != "INSERT INTO `abc_Ext` (`name`,`description`) VALUES ('','')" {
		t.Errorf("build insert failed, str:%v", str)
	}

	str, err = builder.BuildUpdate()
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
	}
	if str != "UPDATE `abc_Ext` SET `name`='',`description`='' WHERE `eid`=12" {
		t.Errorf("build update failed, str:%s", str)
	}

	str, err = builder.BuildDelete()
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
	}
	if str != "DELETE FROM `abc_Ext` WHERE `eid`=12" {
		t.Errorf("build delete failed, str:%s", str)
	}

	uField := extModel.GetField("unit")
	str, err = builder.BuildCreateRelationTable(uField, unitModel)
	if err != nil {
		t.Errorf("BuildCreateRelationTable failed, err:%s", err.Error())
		return
	}
	if str != "CREATE TABLE IF NOT EXISTS `abc_ExtUnit1Unit` (\n\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`left` INT NOT NULL,\n\t`right` TEXT NOT NULL,\n\tPRIMARY KEY (`id`),\n\tINDEX(`left`)\n)\n" {
		t.Errorf("BuildCreateRelationTable failed, str:%s", str)
		return
	}

	str, err = builder.BuildDropRelationTable(uField, unitModel)
	if err != nil {
		t.Errorf("BuildDropRelationTable failed, err:%s", err.Error())
		return
	}
	if str != "DROP TABLE IF EXISTS `abc_ExtUnit1Unit`" {
		t.Errorf("BuildDropRelationTable failed, str:%s", str)
		return
	}

	str, err = builder.BuildInsertRelation(uField, unitModel)
	if err != nil {
		t.Errorf("BuildInsertRelation failed, err:%s", err.Error())
		return
	}
	if str != "INSERT INTO `abc_ExtUnit1Unit` (`left`, `right`) VALUES (12,'10')" {
		t.Errorf("BuildInsertRelation failed, str:%s", str)
		return
	}

	lStr, rStr, err := builder.BuildDeleteRelation(uField, unitModel)
	if err != nil {
		t.Errorf("BuildDeleteRelation failed, err:%s", err.Error())
		return
	}
	if lStr != "DELETE FROM `abc_Unit` WHERE `uid` IN (SELECT `right` FROM `abc_ExtUnit1Unit` WHERE `left`=12)" {
		t.Errorf("BuildDeleteRelation failed, lStr:%s", lStr)
		return
	}
	if rStr != "DELETE FROM `abc_ExtUnit1Unit` WHERE `left`=12" {
		t.Errorf("BuildDeleteRelation failed, rStr:%s", rStr)
		return
	}

	str, err = builder.BuildQueryRelation(uField, unitModel)
	if err != nil {
		t.Errorf("BuildQueryRelation failed, err:%s", err.Error())
		return
	}
	if str != "SELECT `right` FROM `abc_ExtUnit1Unit` WHERE `left`= 12" {
		t.Errorf("BuildQueryRelation failed, str:%s", str)
		return
	}
}

func TestBuilderReference2(t *testing.T) {
	desc := "Desc"
	ext := &Ext{ID: 10, Description: &desc}
	unit := &Unit{ID: "10"}

	localProvider := provider.NewLocalProvider("default")
	_, err := localProvider.RegisterModel(ext)
	if err != nil {
		t.Errorf("localProvider.RegisterModel failed, err:%s", err.Error())
		return
	}
	_, err = localProvider.RegisterModel(unit)
	if err != nil {
		t.Errorf("localProvider.RegisterModel failed, err:%s", err.Error())
		return
	}
	info, err := localProvider.GetEntityModel(ext)
	if err != nil {
		t.Errorf("GetEntityModel failed, err:%s", err.Error())
		return
	}

	builder := NewBuilder(info, localProvider, "abc")
	if builder == nil {
		t.Error("new Builder failed")
	}

	str, err := builder.BuildCreateTable()
	if err != nil {
		t.Errorf("build create schema failed, err:%s", err.Error())
	}
	if str != "CREATE TABLE IF NOT EXISTS `abc_Ext` (\n\t`eid` INT NOT NULL AUTO_INCREMENT,\n\t`name` TEXT NOT NULL ,\n\t`description` TEXT NOT NULL ,\n\tPRIMARY KEY (`eid`)\n)\n" {
		t.Errorf("build create schema failed, str:%v", str)
	}

	str, err = builder.BuildDropTable()
	if err != nil {
		t.Errorf("build drop schema failed, err:%s", err.Error())
	}
	if str != "DROP TABLE IF EXISTS `abc_Ext`" {
		t.Errorf("build drop schema failed, str:%v", str)
	}

	str, err = builder.BuildInsert()
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
	}
	if str != "INSERT INTO `abc_Ext` (`name`,`description`) VALUES ('','Desc')" {
		t.Errorf("build insert failed, str:%v", str)
	}

	str, err = builder.BuildUpdate()
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
	}
	if str != "UPDATE `abc_Ext` SET `name`='',`description`='Desc' WHERE `eid`=10" {
		t.Errorf("build update failed, str:%v", str)
	}

	str, err = builder.BuildDelete()
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
	}
	if str != "DELETE FROM `abc_Ext` WHERE `eid`=10" {
		t.Errorf("build delete failed, str:%v", str)
	}
}
