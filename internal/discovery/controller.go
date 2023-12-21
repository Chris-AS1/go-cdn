package discovery

import (
	"errors"
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
	return catalog[rand.Intn(len(catalog))], nil
}
