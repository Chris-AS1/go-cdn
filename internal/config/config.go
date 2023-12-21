package config

import (
	"fmt"
	"go-cdn/pkg/utils"
	"strings"

	capi "github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
)

func New() (*Config, error) {
	consul_service_id := utils.RandStringBytes(4)
	cfg := Config{
		Consul: Consul{
			ConsulServiceID: consul_service_id,
		},
		Cache:      Cache{RedisEnable: false},
		Database:   Database{DatabaseSSL: false},
		HTTPServer: HTTPServer{DeliveryPort: 3000, RateLimitEnable: false},
		Telemetry:  Telemetry{Sampling: 1, LogPath: "./logs", LogMaxSize: 500, LogMaxBackups: 3, LogMaxAge: 28},
	}

	err := cfg.loadFromFile()
	if cfg.Consul.ConsulServiceAddress == "auto" {
		cfg.Consul.ConsulServiceAddress = utils.GetLocalIPv4()
	}

	return &cfg, err
}

func (cfg *Config) loadFromFile() error {
	viper.SetConfigName("configs")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs/")
	viper.AddConfigPath("/configs/")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("APP") // Allow override from environemnt via APP_VAR_NAME

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	err := viper.Unmarshal(cfg)
	return err
}

func (cfg *Config) GetServiceDefinition() *capi.AgentServiceRegistration {
	csl := cfg.Consul
	return &capi.AgentServiceRegistration{
		ID:      csl.ConsulServiceID,
		Name:    csl.ConsulServiceName,
		Address: csl.ConsulServiceAddress,
		Port:    cfg.HTTPServer.DeliveryPort,
		Check: &capi.AgentServiceCheck{
			Name:                           "web_alive",
			Interval:                       "10s",
			Timeout:                        "30s",
			HTTP:                           fmt.Sprintf("http://%s:%d/health", csl.ConsulServiceAddress, cfg.HTTPServer.DeliveryPort),
			DeregisterCriticalServiceAfter: "1m",
		},
	}
}

func (cfg *Config) GetConsulConfig() *capi.Config {
	return &capi.Config{
		Address:    fmt.Sprintf("%s:%d", cfg.Consul.ConsulAddress, cfg.Consul.ConsulPort),
		Datacenter: cfg.Consul.ConsulDatacenter,
		Scheme:     "http",
	}
}
