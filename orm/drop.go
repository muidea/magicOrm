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

func (s *DropRunner) dropHost() (err *cd.Result) {
	dropResult, dropErr := s.hBuilder.BuildDropTable(s.vModel)
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

func (s *DropRunner) dropRelation(vField model.Field, rModel model.Model) (err *cd.Result) {
	relationResult, relationErr := s.hBuilder.BuildDropRelationTable(s.vModel, vField, rModel)
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
	err = s.dropHost()
	if err != nil {
		log.Errorf("Drop failed, s.dropHost error:%s", err.Error())
		return
	}

	for _, field := range s.vModel.GetFields() {
		if field.IsBasic() {
			continue
		}

		fType := field.GetType()
		rModel, rErr := s.modelProvider.GetTypeModel(fType)
		if rErr != nil {
			err = rErr
			log.Errorf("Drop failed, s.modelProvider.GetTypeModel error:%s", err.Error())
			return
		}

		elemType := fType.Elem()
		if !elemType.IsPtrType() {
			rRunner := NewDropRunner(rModel, s.executor, s.modelProvider, s.modelCodec)
			err = rRunner.Drop()
			if err != nil {
				log.Errorf("Drop failed, rRunner.Drop() error:%s", err.Error())
				return
			}
		}

		err = s.dropRelation(field, rModel)
		if err != nil {
			log.Errorf("dropSchema failed, s.dropRelation error:%s", err.Error())
			return
		}
	}

	return
}

func (s *impl) Drop(vModel model.Model) (err *cd.Result) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	dropRunner := NewDropRunner(vModel, s.executor, s.modelProvider, s.modelCodec)
	err = dropRunner.Drop()
	if err != nil {
		log.Errorf("Drop failed, dropRunner.Drop() error:%s", err.Error())
	}
	return
}
