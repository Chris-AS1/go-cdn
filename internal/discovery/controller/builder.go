package discovery

import (
	"go-cdn/internal/config"
	"go-cdn/internal/discovery/repository/consul"
	"go-cdn/internal/discovery/repository/dummy"
)

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
		consul_repo, err := consul.NewConsulRepo(
			cfg.GetConsulConfig(),
			cfg.GetConsulServiceDefinition(),
		)
		if err != nil {
			return nil, err
		}
		b.SetRepo(consul_repo)
	} else {
		b.SetRepo(dummy.NewDummyRepo())
	}
	return b, nil
}

func (b *ControllerBuilder) Build() *Controller {
	return b.con
}
