package orm

import (
	"context"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"log/slog"
)

type baseRunner struct {
	context       context.Context
	vModel        models.Model
	executor      database.Executor
	modelProvider provider.Provider
	modelCodec    codec.Codec
	sqlBuilder    database.Builder

	batchFilter bool
	deepLevel   int
}

func newBaseRunner(
	ctx context.Context,
	vModel models.Model,
	executor database.Executor,
	provider provider.Provider,
	modelCodec codec.Codec,
	batchFilter bool,
	deepLevel int) baseRunner {
	return baseRunner{
		context:       ctx,
		vModel:        vModel,
		batchFilter:   batchFilter,
		deepLevel:     deepLevel,
		executor:      executor,
		modelProvider: provider,
		modelCodec:    modelCodec,
		sqlBuilder:    NewBuilder(provider, modelCodec),
	}
}

// isContextValid 检查 context 是否失效
func isContextValid(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	select {
	case <-ctx.Done():
		return false
	default:
		return true
	}
}

// checkContext 检查 context 是否失效，如果失效则返回错误
func (s *baseRunner) checkContext() *cd.Error {
	if !isContextValid(s.context) {
		slog.Error("message")
		return cd.NewError(cd.Unexpected, "context is invalid or cancelled")
	}
	return nil
}

// CheckContext 检查 context 是否失效，如果失效则返回错误
func (s *impl) CheckContext() *cd.Error {
	if !isContextValid(s.context) {
		slog.Error("message")
		return cd.NewError(cd.Unexpected, "context is invalid or cancelled")
	}
	return nil
}
