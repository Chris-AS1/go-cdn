package consul

import (
	"go-cdn/config"
	"testing"
)

func TestGetClient(t *testing.T) {
	cfg := config.NewConfig()
	GetClient(cfg.GetConsulConfig())
}
