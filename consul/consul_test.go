package consul

import (
	"go-cdn/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

var cfg *config.Config

func loadConfigsAndGetClient(t *testing.T) (*config.Config, *ConsulClient) {
	c, err := config.NewConfig()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
    
    // Makes it so there's only 1 instance of it, so that Consul service_id doesn't get regenerated
    if cfg == nil {
        cfg = &c
    }

	consul_client, err := NewConsulClient(cfg)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	return cfg, consul_client
}

func TestConsulGetClient(t *testing.T) {
	cfg, err := config.NewConfig()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = NewConsulClient(&cfg)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}

func TestConsulRegistration(t *testing.T) {
	_, consul_client := loadConfigsAndGetClient(t)
	err := consul_client.RegisterService(cfg)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}

func TestConsulDeregistration(t *testing.T) {
	_, consul_client := loadConfigsAndGetClient(t)
	err := consul_client.DeregisterService(cfg)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}

func TestConsulServiceDiscovery(t *testing.T) {
	_, consul_client := loadConfigsAndGetClient(t)
	// Discovers postgres from Consul
	address, port, err := consul_client.DiscoverService(cfg.DatabaseProvider.DatabaseAddress)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.NotSame(t, address, "")
	assert.NotSame(t, port, 0)
}

