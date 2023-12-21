package discovery

type DummyRepository struct{}

// This implementation is used when the external discovery service is disabled so that addresses are read directly from configs.
func NewDummyRepo() *DummyRepository {
	return &DummyRepository{}
}

func (d *DummyRepository) RegisterService() error { return nil }

func (d *DummyRepository) DeregisterService() error { return nil }

func (d *DummyRepository) DiscoverService(address string) ([]string, error) {
	return []string{address}, nil
}
