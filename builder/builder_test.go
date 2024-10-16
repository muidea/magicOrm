package builder

import (
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/util"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/remote"
)

// Unit 单元信息
type Unit struct {
	ID        string    `orm:"uid key uuid"`
	Name      string    `orm:"name"`
	Value     float64   `orm:"value"`
	TimeStamp time.Time `orm:"ts"`
}

type Reference struct {
	//ID 唯一标示单元
	ID int64 `orm:"eid key auto"`
	// Name 名称
	Name        string  `orm:"name"`
	Value       float64 `orm:"value"`
	Description *string `orm:"description"`

	Unit Unit `orm:"unit"`
}

func TestBuilderLocalUnit(t *testing.T) {
	now, _ := time.ParseInLocation(util.CSTLayout, "2018-01-02 15:04:05", time.Local)
	unit := &Unit{ID: "10", Name: "Hello world", Value: 12.3456, TimeStamp: now}

	localProvider := provider.NewLocalProvider("default")
	info, err := localProvider.RegisterModel(unit)
	if err != nil {
		t.Errorf("localProvider.RegisterModel failed, err:%s", err.Error())
		return
	}

	filter, err := localProvider.GetModelFilter(info, 0)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	buildContext := codec.New(localProvider, "abc")
	builder := NewBuilder(info, buildContext)
	if builder == nil {
		t.Error("new Builder failed")
		return
	}

	str, err := builder.BuildCreateTable()
	if err != nil {
		t.Errorf("build create schema failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "CREATE TABLE IF NOT EXISTS `abc_Unit` (\n\t`uid` VARCHAR(32) NOT NULL ,\n\t`name` TEXT NOT NULL ,\n\t`value` DOUBLE NOT NULL ,\n\t`ts` DATETIME NOT NULL ,\n\tPRIMARY KEY (`uid`)\n)\n" {
		t.Errorf("build create schema failed, str:%s", str)
		return
	}

	str, err = builder.BuildDropTable()
	if err != nil {
		t.Errorf("build drop schema failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "DROP TABLE IF EXISTS `abc_Unit`" {
		t.Error("build drop schema failed")
		return
	}

	str, err = builder.BuildInsert()
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
	}
	if str.SQL() != "INSERT INTO `abc_Unit` (`uid`,`name`,`value`,`ts`) VALUES ('10','Hello world',12.3456,'2018-01-02 15:04:05')" {
		t.Errorf("build insert failed, str:%s", str)
		return
	}

	str, err = builder.BuildUpdate()
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "UPDATE `abc_Unit` SET `name` = 'Hello world',`value` = 12.3456,`ts` = '2018-01-02 15:04:05' WHERE `uid` = '10'" {
		t.Errorf("build update failed, str:%s", str)
		return
	}

	str, err = builder.BuildDelete()
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "DELETE FROM `abc_Unit` WHERE `uid` = '10'" {
		t.Errorf("build delete failed, str:%s", str)
		return
	}

	err = filter.Above("value", 12.23)
	if err != nil {
		t.Errorf("filter.Above failed, err:%s", err.Error())
		return
	}
	str, err = builder.BuildQuery(filter)
	if err != nil {
		t.Errorf("build query failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "SELECT `uid`,`name`,`value`,`ts` FROM `abc_Unit` WHERE `value` > 12.23" {
		t.Errorf("build query failed, str:%s", str)
		return
	}

	str, err = builder.BuildCount(filter)
	if err != nil {
		t.Errorf("build count failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "SELECT COUNT(`uid`) FROM `abc_Unit` WHERE `value` > 12.23" {
		t.Errorf("build count failed, str:%s", str)
		return
	}

	return
}

func TestBuilderLocalReference(t *testing.T) {
	var desc string
	ext := &Reference{ID: 12, Name: "Hey", Description: &desc}
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

	buildContext := codec.New(localProvider, "abc")
	builder := NewBuilder(extModel, buildContext)
	if builder == nil {
		t.Error("new Builder failed")
	}

	str, err := builder.BuildCreateTable()
	if err != nil {
		t.Errorf("build create schema failed, err:%s", err.Error())
	}
	if str.SQL() != "CREATE TABLE IF NOT EXISTS `abc_Reference` (\n\t`eid` BIGINT NOT NULL AUTO_INCREMENT,\n\t`name` TEXT NOT NULL ,\n\t`value` DOUBLE NOT NULL ,\n\t`description` TEXT NOT NULL ,\n\tPRIMARY KEY (`eid`)\n)\n" {
		t.Errorf("build create schema failed, str:%s", str)
	}

	str, err = builder.BuildDropTable()
	if err != nil {
		t.Errorf("build drop schema failed, err:%s", err.Error())
	}
	if str.SQL() != "DROP TABLE IF EXISTS `abc_Reference`" {
		t.Errorf("build drop schema failed, str:%s", str)
	}

	str, err = builder.BuildInsert()
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
	}
	if str.SQL() != "INSERT INTO `abc_Reference` (`name`,`value`,`description`) VALUES ('Hey',0,'')" {
		t.Errorf("build insert failed, str:%v", str)
	}

	str, err = builder.BuildUpdate()
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
	}
	if str.SQL() != "UPDATE `abc_Reference` SET `name` = 'Hey',`value` = 0,`description` = '' WHERE `eid` = 12" {
		t.Errorf("build update failed, str:%s", str)
	}

	str, err = builder.BuildDelete()
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
	}
	if str.SQL() != "DELETE FROM `abc_Reference` WHERE `eid` = 12" {
		t.Errorf("build delete failed, str:%s", str)
	}

	uField := extModel.GetField("unit")
	str, err = builder.BuildCreateRelationTable(uField, unitModel)
	if err != nil {
		t.Errorf("BuildCreateRelationTable failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "CREATE TABLE IF NOT EXISTS `abc_ReferenceUnit1Unit` (\n\t`id` BIGINT NOT NULL AUTO_INCREMENT,\n\t`left` BIGINT NOT NULL,\n\t`right` VARCHAR(32) NOT NULL,\n\tPRIMARY KEY (`id`),\n\tINDEX(`left`)\n)\n" {
		t.Errorf("BuildCreateRelationTable failed, str:%s", str)
		return
	}

	str, err = builder.BuildDropRelationTable(uField, unitModel)
	if err != nil {
		t.Errorf("BuildDropRelationTable failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "DROP TABLE IF EXISTS `abc_ReferenceUnit1Unit`" {
		t.Errorf("BuildDropRelationTable failed, str:%s", str)
		return
	}

	str, err = builder.BuildInsertRelation(uField, unitModel)
	if err != nil {
		t.Errorf("BuildInsertRelation failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "INSERT INTO `abc_ReferenceUnit1Unit` (`left`, `right`) VALUES (12,'10')" {
		t.Errorf("BuildInsertRelation failed, str:%s", str)
		return
	}

	lStr, rStr, err := builder.BuildDeleteRelation(uField, unitModel)
	if err != nil {
		t.Errorf("BuildDeleteRelation failed, err:%s", err.Error())
		return
	}
	if lStr.SQL() != "DELETE FROM `abc_Unit` WHERE `uid` IN (SELECT `right` FROM `abc_ReferenceUnit1Unit` WHERE `left`=12)" {
		t.Errorf("BuildDeleteRelation failed, lStr:%s", lStr)
		return
	}
	if rStr.SQL() != "DELETE FROM `abc_ReferenceUnit1Unit` WHERE `left`=12" {
		t.Errorf("BuildDeleteRelation failed, rStr:%s", rStr)
		return
	}

	str, err = builder.BuildQueryRelation(uField, unitModel)
	if err != nil {
		t.Errorf("BuildQueryRelation failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "SELECT `right` FROM `abc_ReferenceUnit1Unit` WHERE `left`= 12" {
		t.Errorf("BuildQueryRelation failed, str:%s", str)
		return
	}
}

func TestBuilderRemoteUnit(t *testing.T) {
	unitObject := &remote.Object{
		Name:    "unit",
		PkgPath: "/test",
		Fields: []*remote.Field{
			{
				Name: "uid",
				Type: &remote.TypeImpl{
					Name:  "string",
					Value: 113,
				},
				Spec: &remote.SpecImpl{
					PrimaryKey:   true,
					ValueDeclare: model.UUID,
				},
			},
			{
				Name: "name",
				Type: &remote.TypeImpl{
					Name:  "string",
					Value: 113,
				},
			},
			{
				Name: "value",
				Type: &remote.TypeImpl{
					Name:  "float64",
					Value: 112,
				},
			},
			{
				Name: "ts",
				Type: &remote.TypeImpl{
					Name:  "dateTime",
					Value: 114,
				},
			},
		},
	}

	unitObjectValue := &remote.ObjectValue{
		Name:    "unit",
		PkgPath: "/test",
		Fields: []*remote.FieldValue{
			{
				Name:  "uid",
				Value: "10",
			},
			{
				Name:  "name",
				Value: "Hello world",
			},
			{
				Name:  "value",
				Value: 12.3456,
			},
			{
				Name:  "ts",
				Value: "2018-01-02 15:04:05",
			},
		},
	}

	rVal := remote.NewValue(unitObjectValue)
	uModel, uErr := remote.SetModelValue(unitObject, rVal)
	if uErr != nil {
		t.Errorf("remote.SetModelValue failed")
		return
	}

	localProvider := provider.NewRemoteProvider("default")
	info, err := localProvider.RegisterModel(uModel)
	if err != nil {
		t.Errorf("localProvider.RegisterModel failed, err:%s", err.Error())
		return
	}

	filter, err := localProvider.GetModelFilter(info, 0)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	buildContext := codec.New(localProvider, "abc")
	builder := NewBuilder(info, buildContext)
	if builder == nil {
		t.Error("new Builder failed")
		return
	}

	str, err := builder.BuildCreateTable()
	if err != nil {
		t.Errorf("build create schema failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "CREATE TABLE IF NOT EXISTS `abc_Unit` (\n\t`uid` VARCHAR(32) NOT NULL ,\n\t`name` TEXT NOT NULL ,\n\t`value` DOUBLE NOT NULL ,\n\t`ts` DATETIME NOT NULL ,\n\tPRIMARY KEY (`uid`)\n)\n" {
		t.Errorf("build create schema failed, str:%s", str)
		return
	}

	str, err = builder.BuildDropTable()
	if err != nil {
		t.Errorf("build drop schema failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "DROP TABLE IF EXISTS `abc_Unit`" {
		t.Error("build drop schema failed")
		return
	}

	str, err = builder.BuildInsert()
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
	}
	if str.SQL() != "INSERT INTO `abc_Unit` (`uid`,`name`,`value`,`ts`) VALUES ('10','Hello world',12.3456,'2018-01-02 15:04:05')" {
		t.Errorf("build insert failed, str:%s", str)
		return
	}

	str, err = builder.BuildUpdate()
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "UPDATE `abc_Unit` SET `name` = 'Hello world',`value` = 12.3456,`ts` = '2018-01-02 15:04:05' WHERE `uid` = '10'" {
		t.Errorf("build update failed, str:%s", str)
		return
	}

	str, err = builder.BuildDelete()
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "DELETE FROM `abc_Unit` WHERE `uid` = '10'" {
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
	if str.SQL() != "SELECT `uid`,`name`,`value`,`ts` FROM `abc_Unit` WHERE `value` > 12" {
		t.Errorf("build query failed, str:%s", str)
		return
	}

	str, err = builder.BuildCount(filter)
	if err != nil {
		t.Errorf("build count failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "SELECT COUNT(`uid`) FROM `abc_Unit` WHERE `value` > 12" {
		t.Errorf("build count failed, str:%s", str)
		return
	}

	return
}

func TestBuilderRemoteReference(t *testing.T) {
	extObject := &remote.Object{
		Name:    "Reference",
		PkgPath: "/test",
		Fields: []*remote.Field{
			{
				Name: "eid",
				Type: &remote.TypeImpl{
					Name:  "int64",
					Value: 105,
				},
				Spec: &remote.SpecImpl{
					PrimaryKey:   true,
					ValueDeclare: model.AutoIncrement,
					ViewDeclare: []model.ViewDeclare{
						model.FullView,
						model.LiteView,
					},
				},
			},
			{
				Name: "name",
				Type: &remote.TypeImpl{
					Name:  "string",
					Value: 113,
				},
				Spec: &remote.SpecImpl{
					ViewDeclare: []model.ViewDeclare{
						model.FullView,
						model.LiteView,
					},
				},
			},
			{
				Name: "value",
				Type: &remote.TypeImpl{
					Name:  "float64",
					Value: 112,
				},
				Spec: &remote.SpecImpl{
					ViewDeclare: []model.ViewDeclare{
						model.FullView,
						model.LiteView,
					},
				},
			},
			{
				Name: "description",
				Type: &remote.TypeImpl{
					Name:  "string",
					Value: 113,
				},
				Spec: &remote.SpecImpl{
					ViewDeclare: []model.ViewDeclare{
						model.FullView,
						model.LiteView,
					},
				},
			},
			{
				Name: "unit",
				Type: &remote.TypeImpl{
					Name:  "Unit",
					Value: 115,
					IsPtr: false,
					ElemType: &remote.TypeImpl{
						Name:    "Unit",
						PkgPath: "/test",
						Value:   115,
					},
				},
				Spec: &remote.SpecImpl{
					ViewDeclare: []model.ViewDeclare{
						model.FullView,
						model.LiteView,
					},
				},
			},
		},
	}

	extObjectValue := &remote.ObjectValue{
		Name:    "Reference",
		PkgPath: "/test",
		Fields: []*remote.FieldValue{
			{
				Name:  "eid",
				Value: 12,
			},
			{
				Name:  "name",
				Value: "Hey",
			},
		},
	}

	unitObject := &remote.Object{
		Name:    "Unit",
		PkgPath: "/test",
		Fields: []*remote.Field{
			{
				Name: "uid",
				Type: &remote.TypeImpl{
					Name:  "string",
					Value: 113,
				},
				Spec: &remote.SpecImpl{
					PrimaryKey:   true,
					ValueDeclare: model.UUID,
				},
			},
			{
				Name: "name",
				Type: &remote.TypeImpl{
					Name:  "string",
					Value: 113,
				},
			},
		},
	}

	unitObjectValue := &remote.ObjectValue{
		Name:    "Unit",
		PkgPath: "/test",
		Fields: []*remote.FieldValue{
			{
				Name:  "uid",
				Value: "10",
			},
			{
				Name:  "name",
				Value: "Hello world",
			},
		},
	}

	uVal := remote.NewValue(unitObjectValue)
	uModel, uErr := remote.SetModelValue(unitObject, uVal)
	if uErr != nil {
		t.Errorf("remote.SetModelValue failed")
		return
	}

	eVal := remote.NewValue(extObjectValue)
	eModel, eErr := remote.SetModelValue(extObject, eVal)
	if eErr != nil {
		t.Errorf("remote.SetModelValue failed")
		return
	}

	remoteProvider := provider.NewRemoteProvider("default")
	extModel, extErr := remoteProvider.RegisterModel(eModel)
	if extErr != nil {
		t.Errorf("remoteProvider.RegisterModel failed, err:%s", extErr.Error())
		return
	}
	unitModel, unitErr := remoteProvider.RegisterModel(uModel)
	if unitErr != nil {
		t.Errorf("remoteProvider.RegisterModel failed, err:%s", unitErr.Error())
		return
	}

	extFilter, extErr := remoteProvider.GetModelFilter(eModel, model.FullView)
	if extErr != nil {
		t.Errorf("remoteProvider.GetModelFilter failed, err:%s", extErr.Error())
		return
	}

	buildContext := codec.New(remoteProvider, "abc")
	builder := NewBuilder(extFilter.MaskModel(), buildContext)
	if builder == nil {
		t.Error("new Builder failed")
		return
	}

	str, err := builder.BuildCreateTable()
	if err != nil {
		t.Errorf("build create schema failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "CREATE TABLE IF NOT EXISTS `abc_Reference` (\n\t`eid` BIGINT NOT NULL AUTO_INCREMENT,\n\t`name` TEXT NOT NULL ,\n\t`value` DOUBLE NOT NULL ,\n\t`description` TEXT NOT NULL ,\n\tPRIMARY KEY (`eid`)\n)\n" {
		t.Errorf("build create schema failed, str:%s", str)
		return
	}

	str, err = builder.BuildDropTable()
	if err != nil {
		t.Errorf("build drop schema failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "DROP TABLE IF EXISTS `abc_Reference`" {
		t.Errorf("build drop schema failed, str:%s", str)
		return
	}

	str, err = builder.BuildInsert()
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "INSERT INTO `abc_Reference` (`name`,`value`,`description`) VALUES ('Hey',0,'')" {
		t.Errorf("build insert failed, str:%v", str)
		return
	}

	str, err = builder.BuildUpdate()
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "UPDATE `abc_Reference` SET `name` = 'Hey',`value` = 0,`description` = '' WHERE `eid` = 12" {
		t.Errorf("build update failed, str:%s", str)
		return
	}

	str, err = builder.BuildDelete()
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "DELETE FROM `abc_Reference` WHERE `eid` = 12" {
		t.Errorf("build delete failed, str:%s", str)
		return
	}

	str, err = builder.BuildQuery(extFilter)
	if err != nil {
		t.Errorf("build query failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "SELECT `eid`,`name`,`value`,`description` FROM `abc_Reference`" {
		t.Errorf("build query failed, str:%s", str)
		return
	}

	extFilter.Equal("eid", 12)
	str, err = builder.BuildQuery(extFilter)
	if err != nil {
		t.Errorf("build query failed, err:%s", err.Error())
	}
	if str.SQL() != "SELECT `eid`,`name`,`value`,`description` FROM `abc_Reference` WHERE `eid` = 12" {
		t.Errorf("build query failed, str:%s", str)
		return
	}

	uField := extModel.GetField("unit")
	str, err = builder.BuildCreateRelationTable(uField, unitModel)
	if err != nil {
		t.Errorf("BuildCreateRelationTable failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "CREATE TABLE IF NOT EXISTS `abc_ReferenceUnit1Unit` (\n\t`id` BIGINT NOT NULL AUTO_INCREMENT,\n\t`left` BIGINT NOT NULL,\n\t`right` VARCHAR(32) NOT NULL,\n\tPRIMARY KEY (`id`),\n\tINDEX(`left`)\n)\n" {
		t.Errorf("BuildCreateRelationTable failed, str:%s", str)
		return
	}

	str, err = builder.BuildDropRelationTable(uField, unitModel)
	if err != nil {
		t.Errorf("BuildDropRelationTable failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "DROP TABLE IF EXISTS `abc_ReferenceUnit1Unit`" {
		t.Errorf("BuildDropRelationTable failed, str:%s", str)
		return
	}

	str, err = builder.BuildInsertRelation(uField, unitModel)
	if err != nil {
		t.Errorf("BuildInsertRelation failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "INSERT INTO `abc_ReferenceUnit1Unit` (`left`, `right`) VALUES (12,'10')" {
		t.Errorf("BuildInsertRelation failed, str:%s", str)
		return
	}

	lStr, rStr, err := builder.BuildDeleteRelation(uField, unitModel)
	if err != nil {
		t.Errorf("BuildDeleteRelation failed, err:%s", err.Error())
		return
	}
	if lStr.SQL() != "DELETE FROM `abc_Unit` WHERE `uid` IN (SELECT `right` FROM `abc_ReferenceUnit1Unit` WHERE `left`=12)" {
		t.Errorf("BuildDeleteRelation failed, lStr:%s", lStr)
		return
	}
	if rStr.SQL() != "DELETE FROM `abc_ReferenceUnit1Unit` WHERE `left`=12" {
		t.Errorf("BuildDeleteRelation failed, rStr:%s", rStr)
		return
	}

	str, err = builder.BuildQueryRelation(uField, unitModel)
	if err != nil {
		t.Errorf("BuildQueryRelation failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "SELECT `right` FROM `abc_ReferenceUnit1Unit` WHERE `left`= 12" {
		t.Errorf("BuildQueryRelation failed, str:%s", str)
		return
	}
}
