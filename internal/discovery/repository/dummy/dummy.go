package dummy

import "go-cdn/internal/discovery/repository"

type DummyRepository struct{}

// This implementation is used when the external discovery service is disabled so that addresses are read directly from configs.
func NewDummyRepo() *DummyRepository {
	return &DummyRepository{}
}

func (d *DummyRepository) RegisterService() error { return repository.ErrServiceDisabled }

func (d *DummyRepository) DeregisterService() error { return repository.ErrServiceDisabled }

func (d *DummyRepository) DiscoverService(address string) ([]string, error) {
	return []string{address}, nil
}
