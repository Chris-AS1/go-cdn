package consul

import (
	"errors"
	"go-cdn/internal/config"

	capi "github.com/hashicorp/consul/api"
)

type ConsulClient struct {
	client *capi.Client
}

func NewConsulClient(cfg *config.Config) (*ConsulClient, error) {
	c := cfg.GetConsulConfig()
	client, err := capi.NewClient(c)
	return &ConsulClient{client}, err
}

func (csl *ConsulClient) RegisterService(cfg *config.Config) error {
	c := csl.client
	s := c.Agent()
	serviceDefinition := cfg.GetServiceDefinition()
	err := s.ServiceRegister(&serviceDefinition)
	return err
}

func (csl *ConsulClient) DeregisterService(cfg *config.Config) error {
	c := csl.client
	err := c.Agent().ServiceDeregister(cfg.Consul.ConsulServiceID)
	return err
}

func (csl *ConsulClient) DiscoverService(service_id string) (string, int, error) {
	services, _, err := csl.client.Catalog().Service(service_id, "", nil)
	if err != nil {
		return "", -1, err
	}

	for _, s := range services {
		return s.ServiceAddress, s.ServicePort, nil
	}

	return "", -1, errors.New("not found")
}
