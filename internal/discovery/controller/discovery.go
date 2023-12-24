package discovery

import (
	"fmt"
	"go-cdn/internal/discovery/repository"
	"math/rand"
)

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
		return "", repository.ErrServiceNotFound
	}
	if len(catalog) <= 0 {
		return "", fmt.Errorf("service_name=%s: %w", service_name, repository.ErrServiceNotFound)
	}
	return catalog[rand.Intn(len(catalog))], nil
}
