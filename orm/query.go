package orm

import (
	"context"
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

type resultItems []any
type resultItemsList []resultItems

type QueryRunner struct {
	baseRunner
}

func NewQueryRunner(
	ctx context.Context,
	vModel models.Model,
	executor database.Executor,
	provider provider.Provider,
	modelCodec codec.Codec,
	batchFilter bool,
	deepLevel int) *QueryRunner {

	return &QueryRunner{
		baseRunner: newBaseRunner(ctx, vModel, executor, provider, modelCodec, batchFilter, deepLevel),
	}
}

func (s *QueryRunner) innerQuery(vModel models.Model, filter models.Filter) (ret resultItemsList, err *cd.Error) {
	queryResult, queryErr := s.sqlBuilder.BuildQuery(vModel, filter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("innerQuery failed, s.sqlBuilder.BuildQuery error:%s", err.Error())
		return
	}

	_, err = s.executor.Query(queryResult.SQL(), false, queryResult.Args()...)
	if err != nil {
		log.Errorf("innerQuery failed, s.executor.Query error:%s", err.Error())
		return
	}
	defer s.executor.Finish()

	queryList := resultItemsList{}
	for s.executor.Next() {
		itemValues, itemErr := s.sqlBuilder.BuildModuleValueHolder(vModel)
		if itemErr != nil {
			err = itemErr
			log.Errorf("innerQuery failed, getModelFieldsPlaceHolder error:%s", err.Error())
			return
		}
		referenceVal := make([]any, len(itemValues))
		for idx := range itemValues {
			referenceVal[idx] = &itemValues[idx]
		}

		err = s.executor.GetField(referenceVal...)
		if err != nil {
			log.Errorf("innerQuery failed, s.executor.GetField error:%s", err.Error())
			return
		}

		queryList = append(queryList, itemValues)
	}

	ret = queryList
	return
}

func (s *QueryRunner) innerAssign(vModel models.Model, queryVal resultItems, deepLevel int) (ret models.Model, err *cd.Error) {
	offset := 0
	qModel := vModel.Copy(models.OriginView)
	for _, field := range qModel.GetFields() {
		if !models.IsBasicField(field) || !models.IsValidField(field) {
			continue
		}

		// 检查 wo 约束，这些字段在查询时应该被排除
		fSpec := field.GetSpec()
		constraints := fSpec.GetConstraints()
		if constraints != nil && constraints.Has(models.KeyWriteOnly) {
			continue
		}

		err = s.assignBasicField(field, queryVal[offset])
		if err != nil {
			log.Errorf("innerAssign field:%s failed, s.assignBasicField error:%v", field.GetName(), err.Error())
			return
		}
		offset++
	}

	for _, field := range qModel.GetFields() {
		if models.IsBasicField(field) || !models.IsValidField(field) {
			continue
		}
		err = s.assignModelField(qModel, field, deepLevel)
		if err != nil {
			log.Errorf("innerAssign field:%s failed, s.assignModelField error:%v", field.GetName(), err.Error())
			return
		}
	}

	ret = qModel
	return
}

func (s *QueryRunner) assignModelField(vModel models.Model, vField models.Field, deepLevel int) (err *cd.Error) {
	vErr := s.queryRelation(vModel, vField, deepLevel)
	if vErr != nil {
		err = vErr
		log.Errorf("assignModelField field:%s failed, s.queryRelation error:%v", vField.GetName(), err.Error())
		return
	}

	return
}

func (s *QueryRunner) assignBasicField(vField models.Field, val any) (err *cd.Error) {
	if val == nil {
		return
	}

	fVal, fErr := s.modelCodec.ExtractBasicFieldValue(vField, val)
	if fErr != nil {
		err = fErr
		log.Errorf("assignBasicField field:%s failed, s.modelProvider.DecodeValue error:%v", vField.GetName(), err.Error())
		return
	}

	vField.SetValue(fVal)
	return
}

func (s *QueryRunner) queryRelation(vModel models.Model, vField models.Field, deepLevel int) (err *cd.Error) {
	if deepLevel > maxDeepLevel {
		return
	}

	if models.IsSliceField(vField) {
		err = s.querySliceRelation(vModel, vField, deepLevel)
		if err != nil {
			log.Errorf("queryRelation field:%s failed, s.querySliceRelation error:%v", vField.GetName(), err.Error())
		}
		return
	}

	err = s.querySingleRelation(vModel, vField, deepLevel)
	if err != nil {
		log.Errorf("queryRelation field:%s failed, s.querySingleRelation error:%v", vField.GetName(), err.Error())
	}
	return
}

func (s *QueryRunner) querySingleRelation(vModel models.Model, vField models.Field, deepLevel int) (err *cd.Error) {
	valueList, valueErr := s.innerQueryRelationKeys(vModel, vField)
	if valueErr != nil {
		err = valueErr
		log.Errorf("querySingleRelation failed, s.innerQueryRelationKeys error:%v", err.Error())
		return
	}

	vType := vField.GetType()
	valueSize := len(valueList)
	if valueSize == 0 {
		if vType.IsPtrType() {
			return
		}
		log.Warnf("query relation failed, field name:%s", vField.GetName())
		return
	}

	rvErr := s.innerQueryRelationSingleModel(valueList[0], vField, deepLevel)
	if rvErr != nil {
		log.Errorf("querySingleRelation failed, s.innerQueryRelationSingleModel error:%v", rvErr.Error())
		err = rvErr
		return
	}
	return
}

func (s *QueryRunner) querySliceRelation(vModel models.Model, vField models.Field, deepLevel int) (err *cd.Error) {
	valueList, valueErr := s.innerQueryRelationKeys(vModel, vField)
	if valueErr != nil {
		err = valueErr
		log.Errorf("querySliceRelation field:%s failed, s.innerQueryRelationKeys error:%sv", vField.GetName(), err.Error())
		return
	}
	valueSize := len(valueList)
	if valueSize == 0 {
		return
	}

	rModelErr := s.innerQueryRelationSliceModel(valueList, vField, deepLevel)
	if rModelErr != nil {
		err = rModelErr
		log.Errorf("querySliceRelation field:%s failed, s.innerQueryRelationSliceModel error:%sv", vField.GetName(), err.Error())
		return
	}
	return
}

func (s *QueryRunner) innerQueryRelationKeys(vModel models.Model, vField models.Field) (ret resultItems, err *cd.Error) {
	relationResult, relationErr := s.sqlBuilder.BuildQueryRelation(vModel, vField)
	if relationErr != nil {
		err = relationErr
		log.Errorf("innerQueryRelationKeys field:%s failed, sqlBuilder.BuildQueryRelation error:%v", vField.GetName(), err.Error())
		return
	}

	values := resultItems{}
	func() {
		_, err = s.executor.Query(relationResult.SQL(), false, relationResult.Args()...)
		if err != nil {
			log.Errorf("innerQueryRelationKeys field:%s failed, s.executor.Query error:%v", vField.GetName(), err.Error())
			return
		}
		defer s.executor.Finish()

		for s.executor.Next() {
			var idVal any
			err = s.executor.GetField(&idVal)
			if err != nil {
				log.Errorf("innerQueryRelationKeys field:%s failed, s.executor.GetField error:%v", vField.GetName(), err.Error())
				return
			}
			values = append(values, idVal)
		}
	}()

	if err != nil {
		return
	}

	ret = values
	return
}

func (s *QueryRunner) innerQueryRelationSingleModel(id any, vField models.Field, deepLevel int) (err *cd.Error) {
	vField.Reset()
	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		err = rErr
		log.Errorf("innerQueryRelationSingleModel failed, s.modelProvider.GetTypeModel field:%s, id:%v, error:%v", vField.GetType().GetPkgKey(), id, err.Error())
		return
	}

	rVal, rErr := s.modelCodec.ExtractBasicFieldValue(rModel.GetPrimaryField(), id)
	if rErr != nil {
		err = rErr
		log.Errorf("innerQueryRelationSingleModel failed, s.modelCodec.ExtractBasicFieldValue field:%s, id:%v, error:%v", rModel.GetPrimaryField().GetType().GetPkgKey(), id, err.Error())
		return
	}
	rModel.SetPrimaryFieldValue(rVal)
	vFilter, vErr := getModelFilter(rModel, s.modelProvider, s.modelCodec)
	if vErr != nil {
		err = vErr
		log.Errorf("innerQueryRelationSingleModel failed, getModelFilter model:%s, id:%v, error:%v", rModel.GetPkgKey(), id, err.Error())
		return
	}

	rQueryRunner := NewQueryRunner(s.context, vFilter.MaskModel(), s.executor, s.modelProvider, s.modelCodec, false, deepLevel+1)
	queryVal, queryErr := rQueryRunner.Query(vFilter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("innerQueryRelationSingleModel failed, s.querySingle model:%s, id:%v, error:%v", rModel.GetPkgKey(), id, err.Error())
		return
	}
	if len(queryVal) > 1 {
		errMsg := fmt.Sprintf("match more than one model, model:%s, id:%v", rModel.GetPkgKey(), id)
		log.Warnf("innerQueryRelationSingleModel failed, errMsg:%s", errMsg)
		err = cd.NewError(cd.Unexpected, errMsg)
		return
	}

	if len(queryVal) > 0 {
		vField.SetValue(queryVal[0].Interface(vField.GetType().Elem().IsPtrType()))
		return
	}

	if deepLevel < maxDeepLevel {
		// 到这里说明未查询到数据，说明存在数据表之间数据不一致
		// 这种情况下直接返回nil，后续要考虑进行脏数据检测
		log.Warnf("query relation failed, miss relation data, model:%s, id:%v", rModel.GetPkgKey(), id)
	}
	return
}

func (s *QueryRunner) innerQueryRelationSliceModel(ids []any, vField models.Field, deepLevel int) (err *cd.Error) {
	// 这里主动重置，避免VFiled的旧数据干扰
	vField.Reset()
	for _, id := range ids {
		svModel, svErr := s.modelProvider.GetTypeModel(vField.GetType().Elem())
		if svErr != nil {
			err = svErr
			log.Errorf("innerQueryRelationSliceModel failed, s.modelProvider.GetTypeModel error:%v", err.Error())
			return
		}
		rVal, rErr := s.modelCodec.ExtractBasicFieldValue(svModel.GetPrimaryField(), id)
		if rErr != nil {
			err = rErr
			log.Errorf("innerQueryRelationSliceModel failed, s.modelCodec.ExtractBasicFieldValue error:%v", err.Error())
			return
		}
		svModel.SetPrimaryFieldValue(rVal)
		vFilter, vErr := getModelFilter(svModel, s.modelProvider, s.modelCodec)
		if vErr != nil {
			err = vErr
			log.Errorf("innerQueryRelationSliceModel failed, getModelFilter error:%v", err.Error())
			return
		}

		rQueryRunner := NewQueryRunner(s.context, vFilter.MaskModel(), s.executor, s.modelProvider, s.modelCodec, false, deepLevel+1)
		queryVal, queryErr := rQueryRunner.Query(vFilter)
		if queryErr != nil {
			err = queryErr
			log.Errorf("innerQueryRelationSingleModel failed, s.querySingle error:%v", err.Error())
			return
		}

		if len(queryVal) > 0 {
			vField.AppendSliceValue(queryVal[0].Interface(vField.GetType().Elem().IsPtrType()))
		}
	}

	return
}

func (s *QueryRunner) Query(filter models.Filter) (ret []models.Model, err *cd.Error) {
	if err = s.checkContext(); err != nil {
		return
	}

	queryValueList, queryValueErr := s.innerQuery(s.vModel, filter)
	if queryValueErr != nil {
		err = queryValueErr
		log.Errorf("Query failed, s.innerQuery error:%s", err.Error())
		return
	}

	queryCount := len(queryValueList)
	if queryCount == 0 {
		return
	}
	if !s.batchFilter && queryCount > 1 {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("matched model:%s %d items value", s.vModel.GetPkgKey(), queryCount))
		log.Warnf("Query failed, s.innerQuery warning:%s", err.Error())
		return
	}

	sliceValue := []models.Model{}
	for idx := range queryValueList {
		modelVal, modelErr := s.innerAssign(s.vModel, queryValueList[idx], 0)
		if modelErr != nil {
			err = modelErr
			log.Errorf("Query failed, s.innerAssign error:%s", err.Error())
			return
		}

		sliceValue = append(sliceValue, modelVal)
	}

	ret = sliceValue
	return
}

func (s *impl) Query(vModel models.Model) (ret models.Model, err *cd.Error) {
	if err = s.CheckContext(); err != nil {
		return
	}

	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "query model is nil")
		return
	}

	// 这里主动Copy一份出来，是为了避免在查询数据过程中对源数据产生了干扰
	vModel = vModel.Copy(models.OriginView)
	vFilter, vErr := getModelFilter(vModel, s.modelProvider, s.modelCodec)
	if vErr != nil {
		err = vErr
		log.Errorf("Query failed, s.getModelFilter error:%v", err.Error())
		return
	}

	vQueryRunner := NewQueryRunner(s.context, vFilter.MaskModel(), s.executor, s.modelProvider, s.modelCodec, false, 0)
	queryVal, queryErr := vQueryRunner.Query(vFilter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("Query failed, vQueryRunner.Query error:%v", err.Error())
		return
	}
	if len(queryVal) != 0 {
		ret = queryVal[0]
		return
	}

	err = cd.NewError(cd.NotFound, fmt.Sprintf("no records found matching the model criteria, model pkgKey: %s, filter: %v", vModel.GetPkgKey(), vFilter))
	return
}
