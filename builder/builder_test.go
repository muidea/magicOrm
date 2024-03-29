package builder

import (
	"testing"
	"time"

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
	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	unit := &Unit{ID: 10, Name: "Hello world", Value: 12.3456, TimeStamp: now}

	provider := provider.NewLocalProvider("default")
	provider.RegisterModel(unit)

	info, err := provider.GetEntityModel(unit)
	if err != nil {
		t.Errorf("GetEntityModel failed, err:%s", err.Error())
		return
	}

	builder := NewBuilder(info, provider)
	if builder == nil {
		t.Error("new Builder failed")
		return
	}

	str, err := builder.BuildCreateSchema()
	if err != nil {
		t.Errorf("build create schema failed, err:%s", err.Error())
		return
	}
	if str != "CREATE TABLE `Unit` (\n\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`name` TEXT NOT NULL ,\n\t`value` DOUBLE NOT NULL ,\n\t`ts` DATETIME NOT NULL ,\n\tPRIMARY KEY (`id`)\n)\n" {
		t.Errorf("build create schema failed, str:%s", str)
		return
	}

	str, err = builder.BuildDropSchema()
	if err != nil {
		t.Errorf("build drop schema failed, err:%s", err.Error())
		return
	}
	if str != "DROP TABLE IF EXISTS `Unit`" {
		t.Error("build drop schema failed")
		return
	}

	str, err = builder.BuildInsert()
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
	}
	if str != "INSERT INTO `Unit` (`name`,`value`,`ts`) VALUES ('Hello world',12.3456,'2018-01-02 15:04:05')" {
		t.Errorf("build insert failed, str:%s", str)
		return
	}

	str, err = builder.BuildUpdate()
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
		return
	}
	if str != "UPDATE `Unit` SET `name`='Hello world',`value`=12.3456,`ts`='2018-01-02 15:04:05' WHERE `id`=10" {
		t.Errorf("build update failed, str:%s", str)
		return
	}

	str, err = builder.BuildDelete()
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
		return
	}
	if str != "DELETE FROM `Unit` WHERE `id`=10" {
		t.Error("build delete failed")
		return
	}

	str, err = builder.BuildQuery(nil)
	if err != nil {
		t.Errorf("build query failed, err:%s", err.Error())
		return
	}
	if str != "SELECT `id`,`name`,`value`,`ts` FROM `Unit`" {
		t.Errorf("build query failed, str:%s", str)
		return
	}

	return
}

func TestBuilderReference(t *testing.T) {
	ext := &Ext{}
	unit := &Unit{}

	provider := provider.NewLocalProvider("default")
	provider.RegisterModel(ext)
	provider.RegisterModel(unit)
	info, err := provider.GetEntityModel(ext)
	if err != nil {
		t.Errorf("GetEntityModel failed, err:%s", err.Error())
		return
	}

	builder := NewBuilder(info, provider)
	if builder == nil {
		t.Error("new Builder failed")
	}

	str, err := builder.BuildCreateSchema()
	if err != nil {
		t.Errorf("build create schema failed, err:%s", err.Error())
	}
	if str != "CREATE TABLE `Ext` (\n\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`name` TEXT NOT NULL ,\n\t`description` TEXT  ,\n\tPRIMARY KEY (`id`)\n)\n" {
		t.Error("build create schema failed")
	}

	str, err = builder.BuildDropSchema()
	if err != nil {
		t.Errorf("build drop schema failed, err:%s", err.Error())
	}
	if str != "DROP TABLE IF EXISTS `Ext`" {
		t.Error("build drop schema failed")
	}

	str, err = builder.BuildInsert()
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
	}
	if str != "INSERT INTO `Ext` (`name`) VALUES ('')" {
		t.Error("build insert failed")
	}

	str, err = builder.BuildUpdate()
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
	}
	if str != "UPDATE `Ext` SET `name`='' WHERE `id`=0" {
		t.Error("build update failed")
	}

	str, err = builder.BuildDelete()
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
	}
	if str != "DELETE FROM `Ext` WHERE `id`=0" {
		t.Error("build delete failed")
	}

	str, err = builder.BuildQuery(nil)
	if err != nil {
		t.Errorf("build query failed, err:%s", err.Error())
	}
	if str != "SELECT `id`,`name`,`description` FROM `Ext`" {
		t.Errorf("build query failed, str:%s", str)
	}
}

func TestBuilderReference2(t *testing.T) {
	desc := "Desc"
	ext := &Ext{Description: &desc}
	unit := &Unit{}

	provider := provider.NewLocalProvider("default")
	provider.RegisterModel(ext)
	provider.RegisterModel(unit)
	info, err := provider.GetEntityModel(ext)
	if err != nil {
		t.Errorf("GetEntityModel failed, err:%s", err.Error())
		return
	}

	builder := NewBuilder(info, provider)
	if builder == nil {
		t.Error("new Builder failed")
	}

	str, err := builder.BuildCreateSchema()
	if err != nil {
		t.Errorf("build create schema failed, err:%s", err.Error())
	}
	if str != "CREATE TABLE `Ext` (\n\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`name` TEXT NOT NULL ,\n\t`description` TEXT  ,\n\tPRIMARY KEY (`id`)\n)\n" {
		t.Error("build create schema failed")
	}

	str, err = builder.BuildDropSchema()
	if err != nil {
		t.Errorf("build drop schema failed, err:%s", err.Error())
	}
	if str != "DROP TABLE IF EXISTS `Ext`" {
		t.Error("build drop schema failed")
	}

	str, err = builder.BuildInsert()
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
	}
	if str != "INSERT INTO `Ext` (`name`,`description`) VALUES ('','Desc')" {
		t.Error("build insert failed")
	}

	str, err = builder.BuildUpdate()
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
	}
	if str != "UPDATE `Ext` SET `name`='',`description`='Desc' WHERE `id`=0" {
		t.Error("build update failed")
	}

	str, err = builder.BuildDelete()
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
	}
	if str != "DELETE FROM `Ext` WHERE `id`=0" {
		t.Error("build delete failed")
	}

	str, err = builder.BuildQuery(nil)
	if err != nil {
		t.Errorf("build query failed, err:%s", err.Error())
	}
	if str != "SELECT `id`,`name`,`description` FROM `Ext`" {
		t.Errorf("build query failed, str:%s", str)
	}
}
