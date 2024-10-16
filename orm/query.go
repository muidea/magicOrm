package orm

import (
	"fmt"

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

func (s *QueryRunner) innerQuery(filter model.Filter) (ret resultItemsList, err *cd.Result) {
	queryResult, queryErr := s.baseRunner.hBuilder.BuildQuery(filter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("innerQuery failed, s.hBuilder.BuildQuery error:%s", err.Error())
		return
	}

	_, err = s.baseRunner.executor.Query(queryResult.SQL(), false, queryResult.Args()...)
	if err != nil {
		log.Errorf("innerQuery failed, s.executor.Query error:%s", err.Error())
		return
	}
	defer s.baseRunner.executor.Finish()

	queryList := resultItemsList{}
	for s.baseRunner.executor.Next() {
		itemValues, itemErr := s.baseRunner.hBuilder.BuildQueryPlaceHolder()
		if itemErr != nil {
			err = itemErr
			if err.Fail() {
				log.Errorf("innerQuery failed, getModelFieldsPlaceHolder error:%s", err.Error())
			} else if err.Warn() {
				log.Warnf("innerQuery failed, getModelFieldsPlaceHolder error:%s", err.Error())
			}
			return
		}

		err = s.baseRunner.executor.GetField(itemValues...)
		if err != nil {
			log.Errorf("innerQuery failed, s.executor.GetField error:%s", err.Error())
			return
		}

		queryList = append(queryList, itemValues)
	}

	ret = queryList
	return
}

func (s *QueryRunner) innerAssign(queryVal resultItems, deepLevel int) (ret model.Model, err *cd.Result) {
	offset := 0

	qModel := s.baseRunner.vModel.Copy(false)
	for _, field := range qModel.GetFields() {
		fValue := field.GetValue()
		if !fValue.IsValid() {
			continue
		}

		if !field.IsBasic() {
			err = s.assignModelField(field, deepLevel)
			if err != nil {
				if err.Fail() {
					log.Errorf("innerAssign failed, s.assignModelField error:%v", err.Error())
				} else if err.Warn() {
					log.Warnf("innerAssign failed, s.assignModelField error:%v", err.Error())
				}
				return
			}

			continue
		}

		err = s.assignBasicField(field, queryVal[offset])
		if err != nil {
			if err.Fail() {
				log.Errorf("innerAssign failed, s.assignBasicField error:%v", err.Error())
			} else if err.Warn() {
				log.Warnf("innerAssign failed, s.assignBasicField error:%v", err.Error())
			}
			return
		}
		offset++
	}

	ret = qModel
	return
}

func (s *QueryRunner) assignModelField(vField model.Field, deepLevel int) (err *cd.Result) {
	vVal, vErr := s.queryRelation(vField, deepLevel)
	if vErr != nil {
		err = vErr
		if err.Fail() {
			log.Errorf("assignModelField failed, s.queryRelation error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("assignModelField failed, s.queryRelation error:%v", err.Error())
		}
		return
	}

	vField.SetValue(vVal)
	return
}

func (s *QueryRunner) assignBasicField(vField model.Field, val any) (err *cd.Result) {
	fVal, fErr := s.baseRunner.modelCodec.ExtractFiledValue(vField, model.NewRawVal(val))
	if fErr != nil {
		err = fErr
		log.Errorf("assignBasicField failed, s.modelProvider.DecodeValue error:%v", err.Error())
		return
	}

	vField.SetValue(fVal)
	return
}

func (s *QueryRunner) queryRelation(vField model.Field, deepLevel int) (ret model.Value, err *cd.Result) {
	if deepLevel > maxDeepLevel {
		fieldType := vField.GetType()
		ret, _ = fieldType.Interface(nil)
		return
	}

	if vField.IsSlice() {
		ret, err = s.querySliceRelation(vField, deepLevel)
		if err != nil {
			if err.Fail() {
				log.Errorf("queryRelation failed, s.querySliceRelation error:%v", err.Error())
			} else if err.Warn() {
				log.Warnf("queryRelation failed, s.querySliceRelation error:%v", err.Error())
			}
		}
		return
	}

	ret, err = s.querySingleRelation(vField, deepLevel)
	if err != nil {
		if err.Fail() {
			log.Errorf("queryRelation failed, s.querySingleRelation error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("queryRelation failed, s.querySingleRelation error:%v", err.Error())
		}
	}
	return
}

func (s *QueryRunner) querySingleRelation(vField model.Field, deepLevel int) (ret model.Value, err *cd.Result) {
	rModel, rErr := s.baseRunner.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		err = rErr
		log.Errorf("querySingleRelation failed, s.modelProvider.GetTypeModel error:%v", err.Error())
		return
	}

	valueList, valueErr := s.innerQueryRelationKeys(vField, rModel)
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
	vValue, _ := vType.Interface(nil)
	valueSize := len(valueList)
	if valueSize == 0 {
		ret = vValue
		if vType.IsPtrType() {
			return
		}

		log.Warnf("query relation failed, field name:%s", vField.GetName())
		//err = cd.NewWarn(cd.Warned, fmt.Sprintf("mismatch relation field:%s", vField.GetName()))
		return
	}

	rvModel, rvErr := s.innerQueryRelationModel(valueList[0], rModel, deepLevel)
	if rvErr != nil {
		if rvErr.Fail() {
			log.Errorf("querySingleRelation failed, s.innerQueryRelationModel error:%v", rvErr.Error())
		} else if rvErr.Warn() {
			log.Warnf("querySingleRelation failed, s.innerQueryRelationModel error:%v", rvErr.Error())
		}

		err = rvErr
		//ret = vValue
		return
	}

	modelVal, modelErr := s.baseRunner.modelProvider.GetEntityValue(rvModel.Interface(vType.IsPtrType(), model.LiteView))
	if modelErr != nil {
		err = modelErr
		log.Errorf("querySingleRelation failed, s.modelProvider.GetEntityValue error:%v", err.Error())
		return
	}

	vValue.Set(modelVal.Get())
	ret = vValue
	return
}

func (s *QueryRunner) querySliceRelation(vField model.Field, deepLevel int) (ret model.Value, err *cd.Result) {
	rModel, rErr := s.baseRunner.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		err = rErr
		log.Errorf("querySliceRelation failed, s.modelProvider.GetTypeModel error:%sv", err.Error())
		return
	}

	vType := vField.GetType()
	vValue, _ := vType.Interface(nil)
	valueList, valueErr := s.innerQueryRelationKeys(vField, rModel)
	if valueErr != nil {
		err = valueErr
		if err.Fail() {
			log.Errorf("querySliceRelation failed, s.innerQueryRelationKeys error:%sv", err.Error())
		} else if err.Warn() {
			log.Warnf("querySliceRelation failed, s.innerQueryRelationKeys error:%sv", err.Error())
		}
		return
	}
	valueSize := len(valueList)
	if valueSize == 0 {
		ret = vValue
		return
	}

	rModelList, rModelErr := s.innerQueryRelationSliceModel(valueList, rModel, deepLevel)
	if rModelErr != nil {
		err = rModelErr
		if err.Fail() {
			log.Errorf("querySliceRelation failed, s.innerQueryRelationSliceModel error:%sv", err.Error())
		} else if err.Warn() {
			log.Warnf("querySliceRelation failed, s.innerQueryRelationSliceModel error:%sv", err.Error())
		}
		return
	}

	elemType := vType.Elem()
	for _, sv := range rModelList {
		modelVal, modelErr := s.baseRunner.modelProvider.GetEntityValue(sv.Interface(elemType.IsPtrType(), model.LiteView))
		if modelErr != nil {
			err = modelErr
			log.Errorf("querySliceRelation failed, s.modelProvider.GetEntityValue error:%sv", err.Error())
			return
		}

		vValue, rErr = s.baseRunner.modelProvider.AppendSliceValue(vValue, modelVal)
		if rErr != nil {
			err = rErr
			log.Errorf("querySliceRelation failed, s.modelProvider.AppendSliceValue error:%sv", err.Error())
			return
		}
	}
	ret = vValue
	return
}

func (s *QueryRunner) innerQueryRelationKeys(vField model.Field, rModel model.Model) (ret resultItems, err *cd.Result) {
	relationResult, relationErr := s.baseRunner.hBuilder.BuildQueryRelation(vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("innerQueryRelationKeys failed, hBuilder.BuildQueryRelation error:%v", err.Error())
		return
	}

	values := resultItems{}
	func() {
		_, err = s.baseRunner.executor.Query(relationResult.SQL(), false, relationResult.Args()...)
		if err != nil {
			log.Errorf("innerQueryRelationKeys failed, s.executor.Query error:%v", err.Error())
			return
		}
		defer s.baseRunner.executor.Finish()

		for s.baseRunner.executor.Next() {
			itemValue, itemErr := s.baseRunner.hBuilder.BuildQueryRelationPlaceHolder(vField, rModel)
			if itemErr != nil {
				err = itemErr
				if err.Fail() {
					log.Errorf("innerQueryRelationKeys failed, s.getModelPKFieldPlaceHolder error:%v", err.Error())
				} else if err.Warn() {
					log.Warnf("innerQueryRelationKeys failed, s.getModelPKFieldPlaceHolder error:%v", err.Error())
				}
				return
			}

			err = s.baseRunner.executor.GetField(itemValue)
			if err != nil {
				log.Errorf("innerQueryRelationKeys failed, s.executor.GetField error:%v", err.Error())
				return
			}
			values = append(values, itemValue)
		}
	}()

	if err != nil {
		return
	}

	ret = values
	return
}

func (s *QueryRunner) innerQueryRelationModel(id any, rModel model.Model, deepLevel int) (ret model.Model, err *cd.Result) {
	pkField := rModel.GetPrimaryField()
	fVal, fErr := s.baseRunner.modelProvider.DecodeValue(model.NewRawVal(id), pkField.GetType())
	if fErr != nil {
		err = fErr
		log.Errorf("innerQueryRelationModel failed, s.modelProvider.DecodeValue error:%v", err.Error())
		return
	}
	rModel.SetFieldValue(pkField.GetName(), fVal)

	vFilter, vErr := getModelFilter(rModel, s.baseRunner.modelProvider, model.LiteView)
	if vErr != nil {
		err = vErr
		if err.Fail() {
			log.Errorf("innerQueryRelationModel failed, getModelFilter error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("innerQueryRelationModel failed, getModelFilter error:%v", err.Error())
		}
		return
	}

	rQueryRunner := NewQueryRunner(rModel, s.baseRunner.executor, s.baseRunner.modelProvider, s.baseRunner.modelCodec, false, deepLevel+1)
	queryVal, queryErr := rQueryRunner.Query(vFilter)
	if queryErr != nil {
		err = queryErr
		if err.Fail() {
			log.Errorf("innerQueryRelationModel failed, s.querySingle error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("innerQueryRelationModel failed, s.querySingle error:%v", err.Error())
		}
		return
	}

	if queryVal != nil && len(queryVal) > 0 {
		ret = queryVal[0]
	}
	return
}

func (s *QueryRunner) innerQueryRelationSliceModel(ids []any, rModel model.Model, deepLevel int) (ret []model.Model, err *cd.Result) {
	sliceVal := []model.Model{}
	for _, id := range ids {
		svModel := rModel.Copy(false)
		pkField := svModel.GetPrimaryField()
		fVal, fErr := s.baseRunner.modelProvider.DecodeValue(model.NewRawVal(id), pkField.GetType())
		if fErr != nil {
			err = fErr
			log.Errorf("innerQueryRelationSliceModel failed, s.modelProvider.DecodeValue error:%v", err.Error())
			return
		}
		svModel.SetFieldValue(pkField.GetName(), fVal)

		vFilter, vErr := getModelFilter(svModel, s.baseRunner.modelProvider, model.LiteView)
		if vErr != nil {
			err = vErr
			if err.Fail() {
				log.Errorf("innerQueryRelationSliceModel failed, getModelFilter error:%v", err.Error())
			} else if err.Warn() {
				log.Warnf("innerQueryRelationSliceModel failed, getModelFilter error:%v", err.Error())
			}
			return
		}

		rQueryRunner := NewQueryRunner(rModel, s.baseRunner.executor, s.baseRunner.modelProvider, s.baseRunner.modelCodec, false, deepLevel+1)
		queryVal, queryErr := rQueryRunner.Query(vFilter)
		if queryErr != nil {
			err = queryErr
			if err.Fail() {
				log.Errorf("innerQueryRelationModel failed, s.querySingle error:%v", err.Error())
			} else if err.Warn() {
				log.Warnf("innerQueryRelationModel failed, s.querySingle error:%v", err.Error())
			}
			return
		}

		if queryVal != nil && len(queryVal) > 0 {
			sliceVal = append(sliceVal, queryVal[0])
		}
	}

	ret = sliceVal
	return
}

func (s *QueryRunner) Query(filter model.Filter) (ret []model.Model, err *cd.Result) {
	queryValueList, queryErr := s.innerQuery(filter)
	if queryErr != nil {
		err = queryErr
		if err.Fail() {
			log.Errorf("Query failed, s.innerQuery error:%s", err.Error())
		} else if err.Warn() {
			log.Warnf("Query failed, s.innerQuery error:%s", err.Error())
		}
		return
	}
	queryCount := len(queryValueList)
	if queryCount == 0 {
		return
	}
	if !s.baseRunner.batchFilter && queryCount > 1 {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("match %d items", queryCount))
		return
	}

	sliceValue := []model.Model{}
	for idx := 0; idx < len(queryValueList); idx++ {
		modelVal, modelErr := s.innerAssign(queryValueList[idx], 0)
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
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	vFilter, vErr := getModelFilter(vModel, s.modelProvider, model.OriginView)
	if vErr != nil {
		err = vErr
		if err.Fail() {
			log.Errorf("Query failed, s.getModelFilter error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("Query failed, s.getModelFilter error:%v", err.Error())
		}
		return
	}

	vQueryRunner := NewQueryRunner(vModel, s.executor, s.modelProvider, s.modelCodec, false, 0)
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

	ret = queryVal[0]
	return
}
