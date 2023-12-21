package discovery

import (
	"fmt"
	"go-cdn/internal/config"

	capi "github.com/hashicorp/consul/api"
)

type ConsulRepository struct {
	client *capi.Client
}

func NewConsulRepo(cfg *config.Config) (*ConsulRepository, error) {
	c := cfg.GetConsulConfig()
	client, err := capi.NewClient(c)
	return &ConsulRepository{client}, err
}

func (csl *ConsulRepository) RegisterService(cfg *config.Config) error {
	c := csl.client
	s := c.Agent()
	serviceDefinition := cfg.GetServiceDefinition()
	err := s.ServiceRegister(&serviceDefinition)
	return err
}

func (csl *ConsulRepository) DeregisterService(cfg *config.Config) error {
	c := csl.client
	err := c.Agent().ServiceDeregister(cfg.Consul.ConsulServiceID)
	return err
}

func (csl *ConsulRepository) DiscoverService(service_name string) (catalog []string, err error) {
	services, _, err := csl.client.Catalog().Service(service_name, "", nil)
	if err != nil {
		return nil, err
	}

	c := []string{}
	for _, s := range services {
		c = append(c, fmt.Sprintf("%s:%d", s.ServiceAddress, s.ServicePort))
	}

	return c, nil
}
