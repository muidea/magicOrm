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
	ID          int64   `orm:"eid key auto"`
	Name        string  `orm:"name"`
	Value       float64 `orm:"value"`
	Description *string `orm:"description"`

	Unit Unit `orm:"unit"`
}

func TestBuilderLocalUnit(t *testing.T) {
	now, _ := time.ParseInLocation(util.CSTLayout, "2018-01-02 15:04:05", time.Local)
	unit := &Unit{ID: "10", Name: "Hello world", Value: 12.3456, TimeStamp: now}

	localProvider := provider.NewLocalProvider("default")
	_, err := localProvider.RegisterModel(unit)
	if err != nil {
		t.Errorf("localProvider.RegisterModel failed, err:%s", err.Error())
	}

	//filter, err := localProvider.GetModelFilter(info)
	//if err != nil {
	//	t.Errorf("GetEntityFilter failed, err:%s", err.Error())
	//	return
	//}

	unitModel, unitErr := localProvider.GetEntityModel(unit)
	if unitErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", unitErr.Error())
	}

	buildContext := codec.New(localProvider, "abc")
	builder := NewBuilder(localProvider, buildContext)
	if builder == nil {
		t.Error("new Builder failed")
	}

	str, err := builder.BuildCreateTable(unitModel)
	if err != nil {
		t.Errorf("build create schema failed, err:%s", err.Error())
	}
	if str.SQL() != "CREATE TABLE IF NOT EXISTS `abc_Unit` (\n\t`uid` VARCHAR(32) NOT NULL,\n\t`name` TEXT NOT NULL,\n\t`value` DOUBLE NOT NULL,\n\t`ts` DATETIME NOT NULL,\n\tPRIMARY KEY (`uid`)\n)\n" {
		t.Errorf("build create schema failed, str:%s", str)
	}

	str, err = builder.BuildDropTable(unitModel)
	if err != nil {
		t.Errorf("build drop schema failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "DROP TABLE IF EXISTS `abc_Unit`" {
		t.Error("build drop schema failed")
	}

	str, err = builder.BuildInsert(unitModel)
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
	}
	if str.SQL() != "INSERT INTO `abc_Unit` (`uid`,`name`,`value`,`ts`) VALUES (?,?,?,?)" || len(str.Args()) != 4 {
		t.Errorf("build insert failed, str:%s", str)
	}

	str, err = builder.BuildUpdate(unitModel)
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
	}
	if str.SQL() != "UPDATE `abc_Unit` SET `name` = ?,`value` = ?,`ts` = ? WHERE `uid` = ?" || len(str.Args()) != 4 {
		t.Errorf("build update failed, str:%s", str)
	}

	str, err = builder.BuildDelete(unitModel)
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
	}
	if str.SQL() != "DELETE FROM `abc_Unit` WHERE `uid` = ?" || len(str.Args()) != 1 {
		t.Errorf("build delete failed, str:%s", str)
	}

	filterModel := unitModel.Copy(model.MetaView)
	filterVal, filterErr := localProvider.GetModelFilter(filterModel)
	if filterErr != nil {
		t.Errorf("GetEntityFilter failed, err:%s", filterErr.Error())
	}

	err = filterVal.Above("value", 12.23)
	if err != nil {
		t.Errorf("filter.Above failed, err:%s", err.Error())
	}
	str, err = builder.BuildQuery(filterModel, filterVal)
	if err != nil {
		t.Errorf("build query failed, err:%s", err.Error())
	}
	if str.SQL() != "SELECT `uid`,`name`,`value`,`ts` FROM `abc_Unit` WHERE `value` > ?" || len(str.Args()) != 1 {
		t.Errorf("build query failed, str:%s", str)
	}

	str, err = builder.BuildCount(filterModel, filterVal)
	if err != nil {
		t.Errorf("build count failed, err:%s", err.Error())
	}
	if str.SQL() != "SELECT COUNT(`uid`) FROM `abc_Unit` WHERE `value` > ?" || len(str.Args()) != 1 {
		t.Errorf("build count failed, str:%s", str)
	}
}

func TestBuilderLocalReference(t *testing.T) {
	var desc string
	unitVal := &Unit{ID: "10"}
	referenceVal := &Reference{
		ID:          12,
		Name:        "Hey",
		Description: &desc,
		Unit: Unit{
			ID: "10",
		},
	}

	localProvider := provider.NewLocalProvider("default")
	referenceModel, referenceErr := localProvider.RegisterModel(referenceVal)
	if referenceErr != nil {
		t.Errorf("localProvider.RegisterModel failed, err:%s", referenceErr.Error())
		return
	}

	unitModel, unitErr := localProvider.RegisterModel(unitVal)
	if unitErr != nil {
		t.Errorf("localProvider.RegisterModel failed, err:%s", unitErr.Error())
	}

	buildContext := codec.New(localProvider, "abc")
	builder := NewBuilder(localProvider, buildContext)
	if builder == nil {
		t.Error("new Builder failed")
	}

	str, err := builder.BuildCreateTable(referenceModel)
	if err != nil {
		t.Errorf("build create schema failed, err:%s", err.Error())
	}
	if str.SQL() != "CREATE TABLE IF NOT EXISTS `abc_Reference` (\n\t`eid` BIGINT NOT NULL AUTO_INCREMENT,\n\t`name` TEXT NOT NULL,\n\t`value` DOUBLE NOT NULL,\n\t`description` TEXT,\n\tPRIMARY KEY (`eid`)\n)\n" {
		t.Errorf("build create schema failed, str:%s", str)
	}

	str, err = builder.BuildDropTable(referenceModel)
	if err != nil {
		t.Errorf("build drop schema failed, err:%s", err.Error())
	}
	if str.SQL() != "DROP TABLE IF EXISTS `abc_Reference`" {
		t.Errorf("build drop schema failed, str:%s", str)
	}

	referenceModel, referenceErr = localProvider.GetEntityModel(referenceVal)
	if referenceErr != nil {
		t.Errorf("localProvider.GetEntityModel failed, err:%s", referenceErr.Error())
		return
	}

	unitModel, unitErr = localProvider.GetEntityModel(unitVal)
	if unitErr != nil {
		t.Errorf("localProvider.GetEntityModel failed, err:%s", unitErr.Error())
		return
	}

	str, err = builder.BuildInsert(referenceModel)
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
	}
	if str.SQL() != "INSERT INTO `abc_Reference` (`name`,`value`,`description`) VALUES (?,?,?)" || len(str.Args()) != 3 {
		t.Errorf("build insert failed, str:%v", str)
	}

	str, err = builder.BuildUpdate(referenceModel)
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
	}
	if str.SQL() != "UPDATE `abc_Reference` SET `name` = ?,`value` = ?,`description` = ? WHERE `eid` = ?" || len(str.Args()) != 4 {
		t.Errorf("build update failed, str:%s", str)
	}

	str, err = builder.BuildDelete(referenceModel)
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
	}
	if str.SQL() != "DELETE FROM `abc_Reference` WHERE `eid` = ?" || len(str.Args()) != 1 {
		t.Errorf("build delete failed, str:%s", str)
	}

	uField := referenceModel.GetField("unit")
	str, err = builder.BuildCreateRelationTable(referenceModel, uField)
	if err != nil {
		t.Errorf("BuildCreateRelationTable failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "CREATE TABLE IF NOT EXISTS `abc_ReferenceUnit1Unit` (\n\t`id` BIGINT NOT NULL AUTO_INCREMENT,\n\t`left` BIGINT NOT NULL,\n\t`right` VARCHAR(32) NOT NULL,\n\tPRIMARY KEY (`id`),\n\tINDEX(`left`)\n)\n" {
		t.Errorf("BuildCreateRelationTable failed, str:%s", str)
		return
	}

	str, err = builder.BuildDropRelationTable(referenceModel, uField)
	if err != nil {
		t.Errorf("BuildDropRelationTable failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "DROP TABLE IF EXISTS `abc_ReferenceUnit1Unit`" {
		t.Errorf("BuildDropRelationTable failed, str:%s", str)
		return
	}

	str, err = builder.BuildInsertRelation(referenceModel, uField, unitModel)
	if err != nil {
		t.Errorf("BuildInsertRelation failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "INSERT INTO `abc_ReferenceUnit1Unit` (`left`, `right`) VALUES (?,?)" || len(str.Args()) != 2 {
		t.Errorf("BuildInsertRelation failed, str:%s", str)
		return
	}

	lStr, rStr, err := builder.BuildDeleteRelation(referenceModel, uField)
	if err != nil {
		t.Errorf("BuildDeleteRelation failed, err:%s", err.Error())
		return
	}
	if lStr.SQL() != "DELETE FROM `abc_Unit` WHERE `uid` IN (SELECT `right` FROM `abc_ReferenceUnit1Unit` WHERE `left`=?)" || len(lStr.Args()) != 1 {
		t.Errorf("BuildDeleteRelation failed, lStr:%s", lStr)
		return
	}
	if rStr.SQL() != "DELETE FROM `abc_ReferenceUnit1Unit` WHERE `left`=?" || len(lStr.Args()) != 1 {
		t.Errorf("BuildDeleteRelation failed, rStr:%s", rStr)
		return
	}

	str, err = builder.BuildQueryRelation(referenceModel, uField)
	if err != nil {
		t.Errorf("BuildQueryRelation failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "SELECT `right` FROM `abc_ReferenceUnit1Unit` WHERE `left`= ?" || len(lStr.Args()) != 1 {
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

	remoteProvider := provider.NewRemoteProvider("default")
	info, err := remoteProvider.RegisterModel(uModel)
	if err != nil {
		t.Errorf("localProvider.RegisterModel failed, err:%s", err.Error())
		return
	}

	filter, err := remoteProvider.GetModelFilter(info)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	buildContext := codec.New(remoteProvider, "abc")
	builder := NewBuilder(remoteProvider, buildContext)
	if builder == nil {
		t.Error("new Builder failed")
		return
	}

	str, err := builder.BuildCreateTable(info)
	if err != nil {
		t.Errorf("build create schema failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "CREATE TABLE IF NOT EXISTS `abc_Unit` (\n\t`uid` VARCHAR(32) NOT NULL,\n\t`name` TEXT NOT NULL,\n\t`value` DOUBLE NOT NULL,\n\t`ts` DATETIME NOT NULL,\n\tPRIMARY KEY (`uid`)\n)\n" {
		t.Errorf("build create schema failed, str:%s", str)
		return
	}

	str, err = builder.BuildDropTable(info)
	if err != nil {
		t.Errorf("build drop schema failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "DROP TABLE IF EXISTS `abc_Unit`" {
		t.Error("build drop schema failed")
		return
	}

	str, err = builder.BuildInsert(info)
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
	}
	if str.SQL() != "INSERT INTO `abc_Unit` (`uid`,`name`,`value`,`ts`) VALUES (?,?,?,?)" || len(str.Args()) != 4 {
		t.Errorf("build insert failed, str:%s", str)
		return
	}

	str, err = builder.BuildUpdate(info)
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "UPDATE `abc_Unit` SET `name` = ?,`value` = ?,`ts` = ? WHERE `uid` = ?" || len(str.Args()) != 4 {
		t.Errorf("build update failed, str:%s", str)
		return
	}

	str, err = builder.BuildDelete(info)
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "DELETE FROM `abc_Unit` WHERE `uid` = ?" || len(str.Args()) != 1 {
		t.Errorf("build delete failed, str:%s", str)
		return
	}

	err = filter.Above("value", 12)
	if err != nil {
		t.Errorf("filter.Above failed, err:%s", err.Error())
		return
	}
	str, err = builder.BuildQuery(info, filter)
	if err != nil {
		t.Errorf("build query failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "SELECT `uid`,`name`,`value`,`ts` FROM `abc_Unit` WHERE `value` > ?" || len(str.Args()) != 1 {
		t.Errorf("build query failed, str:%s", str)
		return
	}

	str, err = builder.BuildCount(info, filter)
	if err != nil {
		t.Errorf("build count failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "SELECT COUNT(`uid`) FROM `abc_Unit` WHERE `value` > ?" || len(str.Args()) != 1 {
		t.Errorf("build count failed, str:%s", str)
		return
	}
}

func TestBuilderRemoteReference(t *testing.T) {
	referenceObject := &remote.Object{
		Name:    "Reference",
		PkgPath: "/test/Reference",
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
						model.DetailView,
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
						model.DetailView,
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
						model.DetailView,
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
						model.DetailView,
						model.LiteView,
					},
				},
			},
			{
				Name: "unit",
				Type: &remote.TypeImpl{
					Name:    "Unit",
					PkgPath: "/test/Unit",
					Value:   115,
					IsPtr:   false,
					ElemType: &remote.TypeImpl{
						Name:    "Unit",
						PkgPath: "/test/Unit",
						Value:   115,
					},
				},
				Spec: &remote.SpecImpl{
					ViewDeclare: []model.ViewDeclare{
						model.DetailView,
						model.LiteView,
					},
				},
			},
		},
	}

	referenceObjectValue := &remote.ObjectValue{
		Name:    "Reference",
		PkgPath: "/test/Reference",
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
		PkgPath: "/test/Unit",
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
		PkgPath: "/test/Unit",
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

	unitModel, uErr := remote.SetModelValue(unitObject, remote.NewValue(unitObjectValue))
	if uErr != nil {
		t.Errorf("remote.SetModelValue failed")
		return
	}

	eVal := remote.NewValue(referenceObjectValue)
	referenceModel, eErr := remote.SetModelValue(referenceObject, eVal)
	if eErr != nil {
		t.Errorf("remote.SetModelValue failed")
		return
	}

	referenceModel.SetFieldValue("unit", unitObjectValue)

	remoteProvider := provider.NewRemoteProvider("default")
	_, extErr := remoteProvider.RegisterModel(referenceModel)
	if extErr != nil {
		t.Errorf("remoteProvider.RegisterModel failed, err:%s", extErr.Error())
		return
	}
	_, unitErr := remoteProvider.RegisterModel(unitModel)
	if unitErr != nil {
		t.Errorf("remoteProvider.RegisterModel failed, err:%s", unitErr.Error())
		return
	}

	extFilter, extErr := remoteProvider.GetModelFilter(referenceModel)
	if extErr != nil {
		t.Errorf("remoteProvider.GetModelFilter failed, err:%s", extErr.Error())
		return
	}

	buildContext := codec.New(remoteProvider, "abc")
	builder := NewBuilder(remoteProvider, buildContext)
	if builder == nil {
		t.Error("new Builder failed")
		return
	}

	str, err := builder.BuildCreateTable(referenceModel)
	if err != nil {
		t.Errorf("build create schema failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "CREATE TABLE IF NOT EXISTS `abc_Reference` (\n\t`eid` BIGINT NOT NULL AUTO_INCREMENT,\n\t`name` TEXT NOT NULL,\n\t`value` DOUBLE NOT NULL,\n\t`description` TEXT NOT NULL,\n\tPRIMARY KEY (`eid`)\n)\n" {
		t.Errorf("build create schema failed, str:%s", str)
		return
	}

	str, err = builder.BuildDropTable(referenceModel)
	if err != nil {
		t.Errorf("build drop schema failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "DROP TABLE IF EXISTS `abc_Reference`" {
		t.Errorf("build drop schema failed, str:%s", str)
		return
	}

	str, err = builder.BuildInsert(referenceModel)
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "INSERT INTO `abc_Reference` (`name`) VALUES (?)" || len(str.Args()) != 1 {
		t.Errorf("build insert failed, str:%v", str)
		return
	}

	str, err = builder.BuildUpdate(referenceModel)
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "UPDATE `abc_Reference` SET `name` = ? WHERE `eid` = ?" || len(str.Args()) != 2 {
		t.Errorf("build update failed, str:%s", str)
		return
	}

	referenceModel.SetFieldValue("value", 12.3456)
	referenceModel.SetFieldValue("description", "Hello world")
	str, err = builder.BuildInsert(referenceModel)
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "INSERT INTO `abc_Reference` (`name`,`value`,`description`) VALUES (?,?,?)" || len(str.Args()) != 3 {
		t.Errorf("build insert failed, str:%v", str)
		return
	}

	str, err = builder.BuildDelete(referenceModel)
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "DELETE FROM `abc_Reference` WHERE `eid` = ?" || len(str.Args()) != 1 {
		t.Errorf("build delete failed, str:%s", str)
		return
	}

	str, err = builder.BuildQuery(extFilter.MaskModel(), extFilter)
	if err != nil {
		t.Errorf("build query failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "SELECT `eid`,`name`,`value`,`description` FROM `abc_Reference`" {
		t.Errorf("build query failed, str:%s", str)
		return
	}

	extFilter.Equal("eid", 12)
	str, err = builder.BuildQuery(extFilter.MaskModel(), extFilter)
	if err != nil {
		t.Errorf("build query failed, err:%s", err.Error())
	}
	if str.SQL() != "SELECT `eid`,`name`,`value`,`description` FROM `abc_Reference` WHERE `eid` = ?" || len(str.Args()) != 1 {
		t.Errorf("build query failed, str:%s", str)
		return
	}

	uField := referenceModel.GetField("unit")
	str, err = builder.BuildCreateRelationTable(referenceModel, uField)
	if err != nil {
		t.Errorf("BuildCreateRelationTable failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "CREATE TABLE IF NOT EXISTS `abc_ReferenceUnit1Unit` (\n\t`id` BIGINT NOT NULL AUTO_INCREMENT,\n\t`left` BIGINT NOT NULL,\n\t`right` VARCHAR(32) NOT NULL,\n\tPRIMARY KEY (`id`),\n\tINDEX(`left`)\n)\n" {
		t.Errorf("BuildCreateRelationTable failed, str:%s", str)
		return
	}

	str, err = builder.BuildDropRelationTable(referenceModel, uField)
	if err != nil {
		t.Errorf("BuildDropRelationTable failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "DROP TABLE IF EXISTS `abc_ReferenceUnit1Unit`" {
		t.Errorf("BuildDropRelationTable failed, str:%s", str)
		return
	}

	str, err = builder.BuildInsertRelation(referenceModel, uField, unitModel)
	if err != nil {
		t.Errorf("BuildInsertRelation failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "INSERT INTO `abc_ReferenceUnit1Unit` (`left`, `right`) VALUES (?,?)" || len(str.Args()) != 2 {
		t.Errorf("BuildInsertRelation failed, str:%s", str)
		return
	}

	lStr, rStr, err := builder.BuildDeleteRelation(referenceModel, uField)
	if err != nil {
		t.Errorf("BuildDeleteRelation failed, err:%s", err.Error())
		return
	}
	if lStr.SQL() != "DELETE FROM `abc_Unit` WHERE `uid` IN (SELECT `right` FROM `abc_ReferenceUnit1Unit` WHERE `left`=?)" || len(lStr.Args()) != 1 {
		t.Errorf("BuildDeleteRelation failed, lStr:%s", lStr)
		return
	}
	if rStr.SQL() != "DELETE FROM `abc_ReferenceUnit1Unit` WHERE `left`=?" || len(rStr.Args()) != 1 {
		t.Errorf("BuildDeleteRelation failed, rStr:%s", rStr)
		return
	}

	str, err = builder.BuildQueryRelation(referenceModel, uField)
	if err != nil {
		t.Errorf("BuildQueryRelation failed, err:%s", err.Error())
		return
	}
	if str.SQL() != "SELECT `right` FROM `abc_ReferenceUnit1Unit` WHERE `left`= ?" || len(str.Args()) != 1 {
		t.Errorf("BuildQueryRelation failed, str:%s", str)
		return
	}
}
