package discovery

import (
	"go-cdn/internal/config"
	"go-cdn/internal/discovery"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsul(t *testing.T) {
	cfg, err := config.New()
	assert.Nil(t, err)

	dc, err := discovery.BuildControllerFromConfigs(cfg)
	assert.Nil(t, err)

	// Don't even attempt to run other tests if a client isn't returned
	if err != nil {
		return
	}

	t.Run("TestConsulRegistration", func(t *testing.T) {
		err := dc.RegisterService()
		assert.Nil(t, err)
	})

	// Looks for itself after registering
	t.Run("TestConsulServiceDiscovery", func(t *testing.T) {
		full_address, err := dc.DiscoverService(cfg.Consul.ConsulServiceName)
		spl_full_address := strings.Split(full_address, ":")
		address, port := spl_full_address[0], spl_full_address[1]
		assert.Nil(t, err)
		assert.NotSame(t, address, "")
		assert.NotSame(t, port, 0)
	})

	t.Run("TestConsulDeregistration", func(t *testing.T) {
		err := dc.DeregisterService()
		assert.Nil(t, err)
	})
}
