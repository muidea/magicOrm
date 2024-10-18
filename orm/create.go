package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

type CreateRunner struct {
	baseRunner
}

func NewCreateRunner(vModel model.Model, executor executor.Executor, provider provider.Provider, modelCodec codec.Codec) *CreateRunner {
	return &CreateRunner{
		baseRunner: newBaseRunner(vModel, executor, provider, modelCodec, false, 0),
	}
}

func (s *CreateRunner) createHost() (err *cd.Result) {
	createResult, createErr := s.hBuilder.BuildCreateTable(s.vModel)
	if createErr != nil {
		err = createErr
		log.Errorf("createHost failed, s.hBuilder.BuildCreateTable error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(createResult.SQL(), createResult.Args()...)
	if err != nil {
		log.Errorf("createHost failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *CreateRunner) createRelation(vField model.Field, rModel model.Model) (err *cd.Result) {
	relationResult, relationErr := s.hBuilder.BuildCreateRelationTable(s.vModel, vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("createRelation failed, hBuilder.BuildCreateRelationTable error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(relationResult.SQL(), relationResult.Args()...)
	if err != nil {
		log.Errorf("createRelation failed, s.executor.Execute error:%s", err.Error())
	}
	return
}

func (s *CreateRunner) Create() (err *cd.Result) {
	err = s.createHost()
	if err != nil {
		log.Errorf("Create failed, s.createHost error:%s", err.Error())
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
			log.Errorf("Create failed, s.modelProvider.GetTypeModel error:%s", err.Error())
			return
		}

		elemType := fType.Elem()
		if !elemType.IsPtrType() {
			rRunner := NewCreateRunner(rModel, s.executor, s.modelProvider, s.modelCodec)
			err = rRunner.Create()
			if err != nil {
				log.Errorf("Create failed, rRunner.Create() error:%s", err.Error())
				return
			}
		}

		err = s.createRelation(field, rModel)
		if err != nil {
			log.Errorf("createSchema failed, s.createRelation error:%s", err.Error())
			return
		}
	}

	return
}

func (s *impl) Create(vModel model.Model) (err *cd.Result) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	createRunner := NewCreateRunner(vModel, s.executor, s.modelProvider, s.modelCodec)
	err = createRunner.Create()
	if err != nil {
		log.Errorf("Create failed, createRunner.Create() error:%s", err.Error())
	}
	return
}
