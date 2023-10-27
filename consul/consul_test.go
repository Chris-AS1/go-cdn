package consul

import (
	"go-cdn/config"
	"testing"
)

func TestConsulGetClient(t *testing.T) {
	cfg := config.NewConfig()
	GetClient(cfg.GetConsulConfig())
}

func TestConsulRegistration(t *testing.T) {
	cfg := config.NewConfig()
	client := GetClient(cfg.GetConsulConfig())
	RegisterService(client, cfg)
}

func TestConsulDeregistration(t *testing.T) {
	cfg := config.NewConfig()
	client := GetClient(cfg.GetConsulConfig())
	DeregisterService(client, cfg)
}
