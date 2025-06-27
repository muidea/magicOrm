package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

type baseRunner struct {
	vModel        model.Model
	executor      executor.Executor
	modelProvider provider.Provider
	modelCodec    codec.Codec
	hBuilder      builder.Builder

	batchFilter bool
	deepLevel   int
}

func newBaseRunner(
	vModel model.Model,
	executor executor.Executor,
	provider provider.Provider,
	modelCodec codec.Codec,
	batchFilter bool,
	deepLevel int) baseRunner {
	return baseRunner{
		vModel:        vModel,
		batchFilter:   batchFilter,
		deepLevel:     deepLevel,
		executor:      executor,
		modelProvider: provider,
		modelCodec:    modelCodec,
		hBuilder:      builder.NewBuilder(provider, modelCodec),
	}
}
