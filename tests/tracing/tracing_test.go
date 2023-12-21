package tracing_test

import (
	"context"
	"go-cdn/internal/config"
	"go-cdn/internal/discovery"
	"go-cdn/internal/tracing"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitPipeline(t *testing.T) {
	cfg, err := config.New()
	assert.Nil(t, err)

	dc, err := discovery.BuildControllerFromConfigs(cfg)
	assert.Nil(t, err)

	ctx := context.Background()

	shutdown, err := tracing.InstallExportPipeline(ctx, dc, cfg)
	defer func() {
		err := shutdown(ctx)
		assert.Nil(t, err)
	}()

	assert.Nil(t, err)
}
