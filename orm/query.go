package orm

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

type resultItems []any
type resultItemsList []resultItems

type QueryRunner struct {
	baseRunner
}

func NewQueryRunner(
	vModel model.Model,
	executor executor.Executor,
	provider provider.Provider,
	modelCodec codec.Codec,
	batchFilter bool,
	deepLevel int) *QueryRunner {

	return &QueryRunner{
		baseRunner: newBaseRunner(vModel, executor, provider, modelCodec, batchFilter, deepLevel),
	}
}

func (s *QueryRunner) innerQuery(vModel model.Model, filter model.Filter) (ret resultItemsList, err *cd.Result) {
	queryResult, queryErr := s.hBuilder.BuildQuery(vModel, filter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("innerQuery failed, s.hBuilder.BuildQuery error:%s", err.Error())
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
		itemValues, itemErr := s.hBuilder.BuildQueryPlaceHolder(vModel)
		if itemErr != nil {
			err = itemErr
			if err.Fail() {
				log.Errorf("innerQuery failed, getModelFieldsPlaceHolder error:%s", err.Error())
			} else if err.Warn() {
				log.Warnf("innerQuery failed, getModelFieldsPlaceHolder error:%s", err.Error())
			}
			return
		}

		err = s.executor.GetField(itemValues...)
		if err != nil {
			log.Errorf("innerQuery failed, s.executor.GetField error:%s", err.Error())
			return
		}

		queryList = append(queryList, itemValues)
	}

	ret = queryList
	return
}

func (s *QueryRunner) innerAssign(vModel model.Model, queryVal resultItems, deepLevel int) (ret model.Model, err *cd.Result) {
	offset := 0
	qModel := vModel.Copy(model.OriginView)
	for _, field := range qModel.GetFields() {
		if !model.IsBasicField(field) || !model.IsValidField(field) {
			continue
		}
		err = s.assignBasicField(field, queryVal[offset])
		if err != nil {
			if err.Fail() {
				log.Errorf("innerAssign field:%s failed, s.assignBasicField error:%v", field.GetName(), err.Error())
			} else if err.Warn() {
				log.Warnf("innerAssign field:%s failed, s.assignBasicField error:%v", field.GetName(), err.Error())
			}
			return
		}
		offset++
	}

	for _, field := range qModel.GetFields() {
		if model.IsBasicField(field) || !model.IsValidField(field) {
			continue
		}
		err = s.assignModelField(qModel, field, deepLevel)
		if err != nil {
			if err.Fail() {
				log.Errorf("innerAssign field:%s failed, s.assignModelField error:%v", field.GetName(), err.Error())
			} else if err.Warn() {
				log.Warnf("innerAssign field:%s failed, s.assignModelField error:%v", field.GetName(), err.Error())
			}
			return
		}
	}

	ret = qModel
	return
}

func (s *QueryRunner) assignModelField(vModel model.Model, vField model.Field, deepLevel int) (err *cd.Result) {
	vErr := s.queryRelation(vModel, vField, deepLevel)
	if vErr != nil {
		err = vErr
		if err.Fail() {
			log.Errorf("assignModelField field:%s failed, s.queryRelation error:%v", vField.GetName(), err.Error())
		} else if err.Warn() {
			log.Warnf("assignModelFiel field:%sd failed, s.queryRelation error:%v", vField.GetName(), err.Error())
		}
		return
	}

	return
}

func (s *QueryRunner) assignBasicField(vField model.Field, val any) (err *cd.Result) {
	fVal, fErr := s.modelCodec.ExtractBasicFieldValue(vField, val)
	if fErr != nil {
		err = fErr
		log.Errorf("assignBasicField field:%s failed, s.modelProvider.DecodeValue error:%v", vField.GetName(), err.Error())
		return
	}

	vField.SetValue(fVal)
	return
}

func (s *QueryRunner) queryRelation(vModel model.Model, vField model.Field, deepLevel int) (err *cd.Result) {
	if deepLevel > maxDeepLevel {
		return
	}

	if model.IsSliceField(vField) {
		err = s.querySliceRelation(vModel, vField, deepLevel)
		if err != nil {
			if err.Fail() {
				log.Errorf("queryRelation field:%s failed, s.querySliceRelation error:%v", vField.GetName(), err.Error())
			} else if err.Warn() {
				log.Warnf("queryRelation field:%s failed, s.querySliceRelation error:%v", vField.GetName(), err.Error())
			}
		}
		return
	}

	err = s.querySingleRelation(vModel, vField, deepLevel)
	if err != nil {
		if err.Fail() {
			log.Errorf("queryRelation field:%s failed, s.querySingleRelation error:%v", vField.GetName(), err.Error())
		} else if err.Warn() {
			log.Warnf("queryRelation field:%s failed, s.querySingleRelation error:%v", vField.GetName(), err.Error())
		}
	}
	return
}

func (s *QueryRunner) querySingleRelation(vModel model.Model, vField model.Field, deepLevel int) (err *cd.Result) {
	valueList, valueErr := s.innerQueryRelationKeys(vModel, vField)
	if valueErr != nil {
		err = valueErr
		if err.Fail() {
			log.Errorf("querySingleRelation failed, s.innerQueryRelationKeys error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("querySingleRelation failed, s.innerQueryRelationKeys error:%v", err.Error())
		}
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
		if rvErr.Fail() {
			log.Errorf("querySingleRelation failed, s.innerQueryRelationSingleModel error:%v", rvErr.Error())
		} else if rvErr.Warn() {
			log.Warnf("querySingleRelation failed, s.innerQueryRelationSingleModel error:%v", rvErr.Error())
		}

		err = rvErr
		return
	}
	return
}

func (s *QueryRunner) querySliceRelation(vModel model.Model, vField model.Field, deepLevel int) (err *cd.Result) {
	valueList, valueErr := s.innerQueryRelationKeys(vModel, vField)
	if valueErr != nil {
		err = valueErr
		if err.Fail() {
			log.Errorf("querySliceRelation field:%s failed, s.innerQueryRelationKeys error:%sv", vField.GetName(), err.Error())
		} else if err.Warn() {
			log.Warnf("querySliceRelation field:%s failed, s.innerQueryRelationKeys error:%sv", vField.GetName(), err.Error())
		}
		return
	}
	valueSize := len(valueList)
	if valueSize == 0 {
		return
	}

	rModelErr := s.innerQueryRelationSliceModel(valueList, vField, deepLevel)
	if rModelErr != nil {
		err = rModelErr
		if err.Fail() {
			log.Errorf("querySliceRelation field:%s failed, s.innerQueryRelationSliceModel error:%sv", vField.GetName(), err.Error())
		} else if err.Warn() {
			log.Warnf("querySliceRelation field:%s failed, s.innerQueryRelationSliceModel error:%sv", vField.GetName(), err.Error())
		}
		return
	}
	return
}

func (s *QueryRunner) innerQueryRelationKeys(vModel model.Model, vField model.Field) (ret resultItems, err *cd.Result) {
	relationResult, relationErr := s.hBuilder.BuildQueryRelation(vModel, vField)
	if relationErr != nil {
		err = relationErr
		log.Errorf("innerQueryRelationKeys field:%s failed, hBuilder.BuildQueryRelation error:%v", vField.GetName(), err.Error())
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
			itemValue, itemErr := s.hBuilder.BuildQueryRelationPlaceHolder(vModel, vField)
			if itemErr != nil {
				err = itemErr
				if err.Fail() {
					log.Errorf("innerQueryRelationKeys field:%s failed, s.getModelPKFieldPlaceHolder error:%v", vField.GetName(), err.Error())
				} else if err.Warn() {
					log.Warnf("innerQueryRelationKeys field:%s failed, s.getModelPKFieldPlaceHolder error:%v", vField.GetName(), err.Error())
				}
				return
			}

			err = s.executor.GetField(itemValue)
			if err != nil {
				log.Errorf("innerQueryRelationKeys field:%s failed, s.executor.GetField error:%v", vField.GetName(), err.Error())
				return
			}

			// 这里需要去除指针
			rawVal := reflect.Indirect(reflect.ValueOf(itemValue)).Interface()
			values = append(values, rawVal)
		}
	}()

	if err != nil {
		return
	}

	ret = values
	return
}

func (s *QueryRunner) innerQueryRelationSingleModel(id any, vField model.Field, deepLevel int) (err *cd.Result) {
	vField.Reset()
	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		err = rErr
		log.Errorf("innerQueryRelationSingleModel failed, s.modelProvider.GetTypeModel field:%s, id:%v, error:%v", vField.GetType().GetPkgKey(), id, err.Error())
		return
	}

	rModel.SetPrimaryFieldValue(id)
	vFilter, vErr := getModelFilter(rModel, s.modelProvider, s.modelCodec)
	if vErr != nil {
		err = vErr
		if err.Fail() {
			log.Errorf("innerQueryRelationSingleModel failed, getModelFilter model:%s, id:%v, error:%v", rModel.GetPkgKey(), id, err.Error())
		} else if err.Warn() {
			log.Warnf("innerQueryRelationSingleModel failed, getModelFilter model:%s, id:%v, error:%v", rModel.GetPkgKey(), id, err.Error())
		}
		return
	}

	rQueryRunner := NewQueryRunner(vFilter.MaskModel(), s.executor, s.modelProvider, s.modelCodec, false, deepLevel+1)
	queryVal, queryErr := rQueryRunner.Query(vFilter)
	if queryErr != nil {
		err = queryErr
		if err.Fail() {
			log.Errorf("innerQueryRelationSingleModel failed, s.querySingle model:%s, id:%v, error:%v", rModel.GetPkgKey(), id, err.Error())
		} else if err.Warn() {
			log.Warnf("innerQueryRelationSingleModel failed, s.querySingle model:%s, id:%v, error:%v", rModel.GetPkgKey(), id, err.Error())
		}
		return
	}
	if len(queryVal) > 1 {
		errMsg := fmt.Sprintf("match more than one model, model:%s, id:%v", rModel.GetPkgKey(), id)
		log.Warnf("innerQueryRelationSingleModel failed, errMsg:%s", errMsg)
		err = cd.NewResult(cd.UnExpected, errMsg)
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

func (s *QueryRunner) innerQueryRelationSliceModel(ids []any, vField model.Field, deepLevel int) (err *cd.Result) {
	// 这里主动重置，避免VFiled的旧数据干扰
	vField.Reset()
	for _, id := range ids {
		svModel, svErr := s.modelProvider.GetTypeModel(vField.GetType().Elem())
		if svErr != nil {
			err = svErr
			log.Errorf("innerQueryRelationSliceModel failed, s.modelProvider.GetTypeModel error:%v", err.Error())
			return
		}

		svModel.SetPrimaryFieldValue(id)
		vFilter, vErr := getModelFilter(svModel, s.modelProvider, s.modelCodec)
		if vErr != nil {
			err = vErr
			if err.Fail() {
				log.Errorf("innerQueryRelationSliceModel failed, getModelFilter error:%v", err.Error())
			} else if err.Warn() {
				log.Warnf("innerQueryRelationSliceModel failed, getModelFilter error:%v", err.Error())
			}
			return
		}

		rQueryRunner := NewQueryRunner(vFilter.MaskModel(), s.executor, s.modelProvider, s.modelCodec, false, deepLevel+1)
		queryVal, queryErr := rQueryRunner.Query(vFilter)
		if queryErr != nil {
			err = queryErr
			if err.Fail() {
				log.Errorf("innerQueryRelationSingleModel failed, s.querySingle error:%v", err.Error())
			} else if err.Warn() {
				log.Warnf("innerQueryRelationSingleModel failed, s.querySingle error:%v", err.Error())
			}
			return
		}

		if len(queryVal) > 0 {
			vField.AppendSliceValue(queryVal[0].Interface(vField.GetType().Elem().IsPtrType()))
		}
	}

	return
}

func (s *QueryRunner) Query(filter model.Filter) (ret []model.Model, err *cd.Result) {
	queryValueList, queryValueErr := s.innerQuery(s.vModel, filter)
	if queryValueErr != nil {
		err = queryValueErr
		if err.Fail() {
			log.Errorf("Query failed, s.innerQuery error:%s", err.Error())
		} else if err.Warn() {
			log.Warnf("Query failed, s.innerQuery warning:%s", err.Error())
		}
		return
	}

	queryCount := len(queryValueList)
	if queryCount == 0 {
		return
	}
	if !s.batchFilter && queryCount > 1 {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("matched model:%s %d items value", s.vModel.GetPkgKey(), queryCount))
		log.Warnf("Query failed, s.innerQuery warning:%s", err.Error())
		return
	}

	sliceValue := []model.Model{}
	for idx := 0; idx < len(queryValueList); idx++ {
		modelVal, modelErr := s.innerAssign(s.vModel, queryValueList[idx], 0)
		if modelErr != nil {
			err = modelErr
			if err.Fail() {
				log.Errorf("Query failed, s.innerAssign error:%s", err.Error())
			} else if err.Warn() {
				log.Warnf("Query failed, s.innerAssign error:%s", err.Error())
			}
			return
		}

		sliceValue = append(sliceValue, modelVal)
	}

	ret = sliceValue
	return
}

func (s *impl) Query(vModel model.Model) (ret model.Model, err *cd.Result) {
	if vModel == nil {
		err = cd.NewResult(cd.IllegalParam, "illegal model value")
		return
	}

	// 这里主动Copy一份出来，是为了避免在查询数据过程中对源数据产生了干扰
	vModel = vModel.Copy(model.OriginView)
	vFilter, vErr := getModelFilter(vModel, s.modelProvider, s.modelCodec)
	if vErr != nil {
		err = vErr
		if err.Fail() {
			log.Errorf("Query failed, s.getModelFilter error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("Query failed, s.getModelFilter error:%v", err.Error())
		}
		return
	}

	vQueryRunner := NewQueryRunner(vFilter.MaskModel(), s.executor, s.modelProvider, s.modelCodec, false, 0)
	queryVal, queryErr := vQueryRunner.Query(vFilter)
	if queryErr != nil {
		err = queryErr
		if err.Fail() {
			log.Errorf("Query failed, vQueryRunner.Query error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("Query failed, vQueryRunner.Query error:%v", err.Error())
		}
		return
	}
	if len(queryVal) != 0 {
		ret = queryVal[0]
		return
	}

	err = cd.NewResult(cd.NoExist, fmt.Sprintf("query model failed, model:%s", vModel.GetPkgKey()))
	return
}
