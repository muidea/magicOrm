package postgres

import (
	"bytes"
	"fmt"
	"strings"

	cd "github.com/muidea/magicCommon/def"

	"log/slog"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
)

func (s *Builder) BuildCreateTable(vModel models.Model) (ret database.Result, err *cd.Error) {
	createSQL := ""
	for _, field := range vModel.GetFields() {
		if !models.IsBasicField(field) {
			continue
		}

		infoVal, infoErr := s.declareFieldInfo(field)
		if infoErr != nil {
			err = infoErr
			slog.Error("BuildCreateTable failed", "operation", "declareFieldInfo", "error", err.Error())
			return
		}

		if createSQL == "" {
			createSQL = fmt.Sprintf("\t%s", infoVal)
		} else {
			createSQL = fmt.Sprintf("%s,\n\t%s", createSQL, infoVal)
		}
	}

	pkFieldName := vModel.GetPrimaryField().GetName()
	createSQL = fmt.Sprintf("%s,\n\tPRIMARY KEY (\"%s\")", createSQL, pkFieldName)
	for _, constraint := range models.GetUniqueConstraints(vModel) {
		if constraintErr := constraint.Verify(vModel.GetFields()); constraintErr != nil {
			return nil, constraintErr
		}
		fields := make([]string, 0, len(constraint.Fields))
		for _, field := range constraint.Fields {
			fields = append(fields, fmt.Sprintf("\"%s\"", field))
		}
		createSQL = fmt.Sprintf("%s,\n\tCONSTRAINT \"%s\" UNIQUE (%s)", createSQL, constraint.Name, strings.Join(fields, ", "))
	}

	createSQL = fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (\n%s\n)\n", s.buildCodec.ConstructModelTableName(vModel), createSQL)
	for _, index := range models.GetIndexes(vModel) {
		if indexErr := index.Verify(vModel.GetFields()); indexErr != nil {
			return nil, indexErr
		}
		fields := make([]string, 0, len(index.Fields))
		for _, field := range index.Fields {
			fields = append(fields, fmt.Sprintf("\"%s\"", field))
		}
		createSQL = fmt.Sprintf("%s;\nCREATE INDEX IF NOT EXISTS \"%s\" ON \"%s\" (%s)", createSQL, index.Name, s.buildCodec.ConstructModelTableName(vModel), strings.Join(fields, ", "))
	}
	if traceSQL() {
		slog.Info("[SQL] create", "sql", createSQL)
	}

	ret = NewError(createSQL, nil)
	return
}

// BuildAddColumn creates an idempotent, additive PostgreSQL migration.  It is
// intentionally limited to one declared basic field; relation storage is
// created through BuildCreateRelationTable.
func (s *Builder) BuildAddColumn(vModel models.Model, vField models.Field) (ret database.Result, err *cd.Error) {
	if !models.IsBasicField(vField) {
		return nil, cd.NewError(cd.IllegalParam, "schema add column requires a basic field")
	}
	info, infoErr := s.declareFieldInfo(vField)
	if infoErr != nil {
		return nil, infoErr
	}
	ret = NewError(fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN IF NOT EXISTS %s", s.buildCodec.ConstructModelTableName(vModel), info), nil)
	return
}

// BuildCreateUniqueConstraint creates a declared unique key. PostgreSQL does
// not support ALTER TABLE ADD CONSTRAINT IF NOT EXISTS, so use its catalog to
// make the operation safe to retry after the database DDL succeeds but the
// caller fails before persisting its schema metadata.
func (s *Builder) BuildCreateUniqueConstraint(vModel models.Model, constraint models.UniqueConstraint) (ret database.Result, err *cd.Error) {
	if err = constraint.Verify(vModel.GetFields()); err != nil {
		return nil, err
	}
	tableName := s.buildCodec.ConstructModelTableName(vModel)
	fields := make([]string, 0, len(constraint.Fields))
	for _, field := range constraint.Fields {
		fields = append(fields, fmt.Sprintf("\"%s\"", field))
	}
	ret = NewError(fmt.Sprintf(`DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM pg_constraint
		WHERE conrelid = '"%s"'::regclass AND conname = '%s'
	) THEN
		ALTER TABLE "%s" ADD CONSTRAINT "%s" UNIQUE (%s);
	END IF;
END $$`, tableName, constraint.Name, tableName, constraint.Name, strings.Join(fields, ", ")), nil)
	return
}

func (s *Builder) BuildCreateIndex(vModel models.Model, index models.Index) (ret database.Result, err *cd.Error) {
	if err = index.Verify(vModel.GetFields()); err != nil {
		return nil, err
	}
	fields := make([]string, 0, len(index.Fields))
	for _, field := range index.Fields {
		fields = append(fields, fmt.Sprintf("\"%s\"", field))
	}
	ret = NewError(fmt.Sprintf("CREATE INDEX IF NOT EXISTS \"%s\" ON \"%s\" (%s)", index.Name, s.buildCodec.ConstructModelTableName(vModel), strings.Join(fields, ", ")), nil)
	return
}

// BuildCreateRelationTable Build CreateRelation Schema
func (s *Builder) BuildCreateRelationTable(vModel models.Model, vField models.Field) (ret database.Result, err *cd.Error) {
	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField)
	if relationErr != nil {
		err = relationErr
		slog.Error("BuildCreateRelationTable failed", "field", vField.GetName(), "operation", "s.buildCodec.ConstructRelationTableName", "error", err.Error())
		return
	}

	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType().Elem())
	if rErr != nil {
		err = rErr
		slog.Error("BuildCreateRelationTable failed", "field", vField.GetName(), "operation", "s.modelProvider.GetTypeModel", "error", err.Error())
		return
	}

	lPKField := vModel.GetPrimaryField()
	lPKType, lPKErr := getTypeDeclare(lPKField.GetType(), lPKField.GetSpec(), false)
	if lPKErr != nil {
		err = lPKErr
		slog.Error("BuildCreateRelationTable failed", "field", lPKField.GetName(), "operation", "getTypeDeclare", "error", err.Error())
		return
	}

	rPKField := rModel.GetPrimaryField()
	rPKType, rPKErr := getTypeDeclare(rPKField.GetType(), rPKField.GetSpec(), false)
	if rPKErr != nil {
		err = rPKErr
		slog.Error("BuildCreateRelationTable failed", "field", rPKField.GetName(), "operation", "getTypeDeclare", "error", err.Error())
		return
	}

	createRelationSQL := fmt.Sprintf("\t\"id\" BIGSERIAL NOT NULL,\n\t\"left\" %s NOT NULL,\n\t\"right\" %s NOT NULL,\n\tPRIMARY KEY (\"id\")", lPKType, rPKType)
	createRelationSQL = fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (\n%s\n)", relationTableName, createRelationSQL)
	createRelationSQL = fmt.Sprintf("%s;\nCREATE INDEX IF NOT EXISTS \"%s_index\" ON \"%s\" (\"left\")", createRelationSQL, relationTableName, relationTableName)
	if traceSQL() {
		slog.Info("[SQL] create relation", "sql", createRelationSQL)
	}

	ret = NewError(createRelationSQL, nil)
	return
}

// declareFieldInfo declare field info
// 根据字段类型和字段特性生成字段定义
// 类似以下信息
// "id" int(11) NOT NULL AUTO_INCREMENT
// "i8" tinyint(4) DEFAULT '100',
func (s *Builder) declareFieldInfo(vField models.Field) (ret string, err *cd.Error) {
	strBuffer := bytes.NewBufferString("")
	// Write field name
	strBuffer.WriteString("\"")
	strBuffer.WriteString(vField.GetName())
	strBuffer.WriteString("\"")

	// Write field type
	typeVal, typeErr := getTypeDeclare(vField.GetType(), vField.GetSpec(), true)
	if typeErr != nil {
		err = typeErr
		slog.Error("declareFieldInfo failed", "operation", "getTypeDeclare", "error", err.Error())
		return
	}
	strBuffer.WriteString(" ")
	strBuffer.WriteString(typeVal)

	// Write NULL constraint
	if !vField.GetType().IsPtrType() {
		strBuffer.WriteString(" NOT NULL")
	}

	// Write default value if exists
	fSpec := vField.GetSpec()
	defaultValue, defaultErr := s.validDefaultValue(vField.GetType(), fSpec)
	if defaultErr != nil {
		err = defaultErr
		slog.Error("declareFieldInfo failed", "operation", "validDefaultValue", "error", err.Error())
		return
	}
	// Write auto increment if needed
	autoIncVal, autoIncErr := s.validAutoIncrement(vField.GetType(), vField.GetSpec())
	if autoIncErr != nil {
		err = autoIncErr
		slog.Error("declareFieldInfo failed", "operation", "validAutoIncrement", "error", err.Error())
		return
	}

	if !autoIncVal && defaultValue != "''" && defaultValue != "" {
		strBuffer.WriteString(" DEFAULT ")
		strBuffer.WriteString(defaultValue)
	}

	if autoIncVal {
		// PostgreSQL 使用 SERIAL 类型代替 AUTO_INCREMENT
		// 这里需要修改字段类型而不是添加 AUTO_INCREMENT 关键字
		// 在 getTypeDeclare 中已经处理了 SERIAL 类型
	}

	ret = strBuffer.String()
	return
}

func (s *Builder) validDefaultValue(vType models.Type, vSpec models.Spec) (ret string, err *cd.Error) {
	if !models.IsBasic(vType) || vType.IsPtrType() || vType.GetValue().IsSliceType() || vType.Elem().GetValue().IsStringValueType() {
		// 非基础类型和切片类型不需要设置默认值
		// 指针类型不需要设置默认值
		// 字符串类型不需要设置默认值，这里返回空
		return
	}

	var defaultValue, defaultValueDeclare any
	if vSpec != nil {
		defaultValueDeclare = vSpec.GetDefaultValue()
	}
	if defaultValueDeclare != nil {
		switch val := defaultValueDeclare.(type) {
		case string:
			if strings.HasPrefix(val, "$reference") {
				vTypeDefaultVal, _ := vType.Interface(nil)
				defaultValue = vTypeDefaultVal.Get()
			} else {
				defaultValue = val
			}
		default:
			defaultValue = val
		}
	} else {
		vTypeDefaultVal, _ := vType.Interface(nil)
		defaultValue = vTypeDefaultVal.Get()
	}

	switch vType.Elem().GetValue() {
	case models.TypeBooleanValue:
		if defaultValue.(bool) {
			ret = "'1'"
		} else {
			ret = "'0'"
		}
	default:
		ret = fmt.Sprintf("'%v'", defaultValue)
	}

	return
}

func (s *Builder) validAutoIncrement(vType models.Type, vSpec models.Spec) (ret bool, err *cd.Error) {
	if vSpec == nil || !vSpec.IsPrimaryKey() || !vType.GetValue().IsNumberValueType() {
		return
	}

	ret = models.IsAutoIncrementDeclare(vSpec.GetValueDeclare())
	return
}
