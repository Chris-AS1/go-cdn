package discovery

type DummyRepository struct {
}

func NewDummyRepo() *DummyRepository {
	return &DummyRepository{}
}

func (d *DummyRepository) RegisterService() error { return nil }

func (d *DummyRepository) DeregisterService() error { return nil }

func (d *DummyRepository) DiscoverService(address string) ([]string, error) {
	return []string{address}, nil
}
