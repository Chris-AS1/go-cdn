package discovery

import (
	"errors"
	"go-cdn/internal/config"
	"math/rand"
)

var ErrServiceNotFound = errors.New("service not found")

type discoveryRepository interface {
	RegisterService(*config.Config) error
	DeregisterService(*config.Config) error
	DiscoverService(string) ([]string, error)
}

type Controller struct {
	repo discoveryRepository
}

func NewController(repo discoveryRepository) *Controller {
	return &Controller{repo}
}

func (c *Controller) RegisterService(cfg *config.Config) error {
	err := c.repo.RegisterService(cfg)
	return err
}

func (c *Controller) DeregisterService(cfg *config.Config) error {
	err := c.repo.DeregisterService(cfg)
	return err
}

func (c *Controller) DiscoverService(service_name string) (string, error) {
	catalog, err := c.repo.DiscoverService(service_name)
	if err != nil {
		return "", ErrServiceNotFound
	}
	return catalog[rand.Intn(len(catalog))], nil
}
