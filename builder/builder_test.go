package builder

import (
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/util"

	"github.com/muidea/magicOrm/provider"
)

// Unit 单元信息
type Unit struct {
	ID        int       `orm:"id key auto"`
	Name      string    `orm:"name"`
	Value     float64   `orm:"value"`
	TimeStamp time.Time `orm:"ts"`
}

type Ext struct {
	//ID 唯一标示单元
	ID int `orm:"id key auto"`
	// Name 名称
	Name string `orm:"name"`

	Description *string `orm:"description"`

	Unit Unit `orm:"unit"`
}

func TestBuilderCommon(t *testing.T) {
	now, _ := time.ParseInLocation(util.CSTLayout, "2018-01-02 15:04:05", time.Local)
	unit := &Unit{ID: 10, Name: "Hello world", Value: 12.3456, TimeStamp: now}

	localProvider := provider.NewLocalProvider("default")
	localProvider.RegisterModel(unit)

	info, err := localProvider.GetEntityModel(unit)
	if err != nil {
		t.Errorf("GetEntityModel failed, err:%s", err.Error())
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
	if str != "CREATE TABLE IF NOT EXISTS `abc_Unit` (\n\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`name` TEXT NOT NULL ,\n\t`value` DOUBLE NOT NULL ,\n\t`ts` DATETIME NOT NULL ,\n\tPRIMARY KEY (`id`)\n)\n" {
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
	if str != "INSERT INTO `abc_Unit` (`name`,`value`,`ts`) VALUES ('Hello world',12.3456,'2018-01-02 15:04:05')" {
		t.Errorf("build insert failed, str:%s", str)
		return
	}

	str, err = builder.BuildUpdate()
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
		return
	}
	if str != "UPDATE `abc_Unit` SET `name`='Hello world',`value`=12.3456,`ts`='2018-01-02 15:04:05' WHERE `id`=10" {
		t.Errorf("build update failed, str:%s", str)
		return
	}

	str, err = builder.BuildDelete()
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
		return
	}
	if str != "DELETE FROM `abc_Unit` WHERE `id`=10" {
		t.Error("build delete failed")
		return
	}

	str, err = builder.BuildQuery(nil)
	if err != nil {
		t.Errorf("build query failed, err:%s", err.Error())
		return
	}
	if str != "SELECT `id`,`name`,`value`,`ts` FROM `abc_Unit`" {
		t.Errorf("build query failed, str:%s", str)
		return
	}

	return
}

func TestBuilderReference(t *testing.T) {
	ext := &Ext{ID: 10}
	unit := &Unit{ID: 10}

	localProvider := provider.NewLocalProvider("default")
	localProvider.RegisterModel(ext)
	localProvider.RegisterModel(unit)
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
	if str != "CREATE TABLE IF NOT EXISTS `abc_Ext` (\n\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`name` TEXT NOT NULL ,\n\t`description` TEXT NOT NULL ,\n\tPRIMARY KEY (`id`)\n)\n" {
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
	if str != "UPDATE `abc_Ext` SET `name`='' WHERE `id`=10" {
		t.Errorf("build update failed, str:%s", str)
	}

	str, err = builder.BuildDelete()
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
	}
	if str != "DELETE FROM `abc_Ext` WHERE `id`=10" {
		t.Errorf("build delete failed, str:%s", str)
	}

	str, err = builder.BuildQuery(nil)
	if err != nil {
		t.Errorf("build query failed, err:%s", err.Error())
	}
	if str != "SELECT `id`,`name`,`description` FROM `abc_Ext`" {
		t.Errorf("build query failed, str:%s", str)
	}
}

func TestBuilderReference2(t *testing.T) {
	desc := "Desc"
	ext := &Ext{ID: 10, Description: &desc}
	unit := &Unit{ID: 10}

	localProvider := provider.NewLocalProvider("default")
	localProvider.RegisterModel(ext)
	localProvider.RegisterModel(unit)
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
	if str != "CREATE TABLE IF NOT EXISTS `abc_Ext` (\n\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`name` TEXT NOT NULL ,\n\t`description` TEXT NOT NULL ,\n\tPRIMARY KEY (`id`)\n)\n" {
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
	if str != "UPDATE `abc_Ext` SET `name`='',`description`='Desc' WHERE `id`=10" {
		t.Errorf("build update failed, str:%v", str)
	}

	str, err = builder.BuildDelete()
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
	}
	if str != "DELETE FROM `abc_Ext` WHERE `id`=10" {
		t.Errorf("build delete failed, str:%v", str)
	}

	str, err = builder.BuildQuery(nil)
	if err != nil {
		t.Errorf("build query failed, err:%s", err.Error())
	}
	if str != "SELECT `id`,`name`,`description` FROM `abc_Ext`" {
		t.Errorf("build query failed, str:%s", str)
	}
}
