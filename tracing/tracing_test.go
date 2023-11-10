package tracing

import (
	"context"
	"go-cdn/config"
	"go-cdn/consul"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitPipeline(t *testing.T) {
	cfg, err := config.NewConfig()
	assert.Nil(t, err)

	consul_client, err := consul.NewConsulClient(&cfg)
	assert.Nil(t, err)

	ctx := context.Background()

	shutdown, err := InstallExportPipeline(ctx, consul_client, &cfg)
	defer func() {
		err := shutdown(ctx)
		assert.Nil(t, err)
	}()

	assert.Nil(t, err)
}
