package discovery

import (
	"fmt"

	capi "github.com/hashicorp/consul/api"
)

type ConsulRepository struct {
	client *capi.Client
	sc     *capi.Config
	sd     *capi.AgentServiceRegistration
}

func NewConsulRepo(sc *capi.Config, sd *capi.AgentServiceRegistration) (*ConsulRepository, error) {
	client, err := capi.NewClient(sc)
	return &ConsulRepository{client, sc, sd}, err
}

func (c *ConsulRepository) RegisterService() error {
	client := c.client
	s := client.Agent()
	err := s.ServiceRegister(c.sd)
	return err
}

func (c *ConsulRepository) DeregisterService() error {
	client := c.client
	err := client.Agent().ServiceDeregister(c.sd.ID)
	return err
}

func (c *ConsulRepository) DiscoverService(service_name string) ([]string, error) {
	services, _, err := c.client.Catalog().Service(service_name, "", nil)
	if err != nil {
		return nil, err
	}

	catalog := []string{}
	for _, s := range services {
		catalog = append(catalog, fmt.Sprintf("%s:%d", s.ServiceAddress, s.ServicePort))
	}

	return catalog, nil
}
