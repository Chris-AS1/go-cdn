package discovery

import (
	"errors"
	"fmt"
	"go-cdn/internal/config"
	"math/rand"
)

var ErrServiceNotFound = errors.New("service not found")
var ErrServiceDisabled = errors.New("discovery service is disabled")

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

type ControllerBuilder struct {
	con *Controller
}

func NewControllerBuilder() *ControllerBuilder {
	return &ControllerBuilder{}
}

func (b *ControllerBuilder) SetRepo(repo discoveryRepository) *ControllerBuilder {
	b.con = NewController(repo)
	return b
}

func (b *ControllerBuilder) FromConfigs(cfg *config.Config) (*ControllerBuilder, error) {
	if cfg.Consul.ConsulEnable {
		consul_repo, err := NewConsulRepo(
			cfg.GetConsulConfig(),
			cfg.GetServiceDefinition(),
		)
		if err != nil {
			return nil, err
		}
		b.SetRepo(consul_repo)
	} else {
		b.SetRepo(NewDummyRepo())
	}
	return b, nil
}

func (b *ControllerBuilder) Build() *Controller {
	return b.con
}
