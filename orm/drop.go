package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

type DropRunner struct {
	baseRunner
}

func NewDropRunner(vModel model.Model, executor executor.Executor, provider provider.Provider, modelCodec codec.Codec) *DropRunner {
	return &DropRunner{
		baseRunner: newBaseRunner(vModel, executor, provider, modelCodec, false, 0),
	}
}

func (s *DropRunner) dropHost(vModel model.Model) (err *cd.Result) {
	dropResult, dropErr := s.hBuilder.BuildDropTable(vModel)
	if dropErr != nil {
		err = dropErr
		log.Errorf("dropHost failed, s.hBuilder.BuildDropTable error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(dropResult.SQL(), dropResult.Args()...)
	if err != nil {
		log.Errorf("dropHost failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *DropRunner) dropRelation(vModel model.Model, vField model.Field) (err *cd.Result) {
	relationResult, relationErr := s.hBuilder.BuildDropRelationTable(vModel, vField)
	if relationErr != nil {
		err = relationErr
		log.Errorf("dropRelation failed, hBuilder.BuildDropRelationTable error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(relationResult.SQL(), relationResult.Args()...)
	if err != nil {
		log.Errorf("dropRelation failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *DropRunner) Drop() (err *cd.Result) {
	err = s.dropHost(s.vModel)
	if err != nil {
		log.Errorf("Drop failed, s.dropHost error:%s", err.Error())
		return
	}

	for _, field := range s.vModel.GetFields() {
		if model.IsBasicField(field) {
			continue
		}

		elemType := field.GetType().Elem()
		if !elemType.IsPtrType() {
			rModel, rErr := s.modelProvider.GetTypeModel(elemType)
			if rErr != nil {
				err = rErr
				log.Errorf("Drop failed, s.modelProvider.GetTypeModel error:%s", err.Error())
				return
			}

			rRunner := NewDropRunner(rModel, s.executor, s.modelProvider, s.modelCodec)
			err = rRunner.Drop()
			if err != nil {
				log.Errorf("Drop failed, rRunner.Drop() error:%s", err.Error())
				return
			}
		}

		err = s.dropRelation(s.vModel, field)
		if err != nil {
			log.Errorf("Drop failed, s.dropRelation error:%s", err.Error())
			return
		}
	}

	return
}

func (s *impl) Drop(vModel model.Model) (err *cd.Result) {
	if vModel == nil {
		err = cd.NewResult(cd.IllegalParam, "illegal model value")
		return
	}

	dropRunner := NewDropRunner(vModel, s.executor, s.modelProvider, s.modelCodec)
	err = dropRunner.Drop()
	if err != nil {
		log.Errorf("Drop failed, dropRunner.Drop() error:%s", err.Error())
	}
	return
}
