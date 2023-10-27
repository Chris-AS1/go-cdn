package consul

import (
	"go-cdn/config"
	"log"
	capi "github.com/hashicorp/consul/api"
)

func GetClient(c *capi.Config) *capi.Client {
	client, err := capi.NewClient(c)
	if err != nil {
		panic(err)
	}
	return client
}

func RegisterService(c *capi.Client, cfg config.Config) {
	s := c.Agent()
	serviceDefinition := cfg.GetServiceDefinition()
	if err := s.ServiceRegister(&serviceDefinition); err != nil {
		log.Panic(err)
	}
}

func DeregisterService(c *capi.Client, cfg config.Config) {
	if err := c.Agent().ServiceDeregister(cfg.Consul.ConsulServiceID); err != nil {
		log.Panic(err)
	}
}

