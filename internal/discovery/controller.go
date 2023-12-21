package discovery

import (
	"errors"
	"fmt"
	"go-cdn/internal/config"
	"math/rand"
)

var ErrServiceNotFound = errors.New("service not found")
var ErrServiceDisabled = errors.New("discovery service is disabled")

func BuildControllerFromConfigs(cfg *config.Config) (*Controller, error) {
	if cfg.Consul.ConsulEnable {
		consul_repo, err := NewConsulRepo(
			cfg.GetConsulConfig(),
			cfg.GetServiceDefinition(),
		)
		if err != nil {
			return nil, err
		}

		return NewController(consul_repo), nil
	} else {
		return NewController(NewDummyRepo()), nil
	}
}

type discoveryRepository interface {
	RegisterService() error
	DeregisterService() error
	DiscoverService(string) ([]string, error)
}

type Controller struct {
	repo discoveryRepository
}

func NewController(repo discoveryRepository) *Controller {
	return &Controller{repo}
}

func (c *Controller) RegisterService() error {
	err := c.repo.RegisterService()
	return err
}

func (c *Controller) DeregisterService() error {
	err := c.repo.DeregisterService()
	return err
}

func (c *Controller) DiscoverService(service_name string) (string, error) {
	catalog, err := c.repo.DiscoverService(service_name)
	if err != nil {
		return "", ErrServiceNotFound
	}
	if len(catalog) <= 0 {
		return "", fmt.Errorf("%s: %s", ErrServiceNotFound, service_name)
	}
	return catalog[rand.Intn(len(catalog))], nil
}
