package config

import (
	"fmt"
	"go-cdn/utils"
	"strings"

	capi "github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
)

type Config struct {
	Consul           Consul           `mapstructure:"consul"`
	Redis            RedisDatabase    `mapstructure:"redis"`
	DatabaseProvider DatabaseProvider `mapstructure:"postgres"`
	HTTPServer       HTTPServer       `mapstructure:"http"`
	Telemetry        Telemetry        `mapstructure:"telemetry"`
}

type Consul struct {
	ConsulEnable         bool   `mapstructure:"enable"`
	ConsulServiceID      string `mapstructure:"service_id"` // Should not be present in configs.yaml. It's randomized for each instance
	ConsulServiceName    string `mapstructure:"service_name"`
	ConsulServiceAddress string `mapstructure:"service_address"`
	ConsulAddress        string `mapstructure:"address"`
	ConsulDatacenter     string `mapstructure:"datacenter"`
	ConsulPort           int    `mapstructure:"port"`
}

type RedisDatabase struct {
	RedisEnable   bool   `mapstructure:"enable"`
	RedisAddress  string `mapstructure:"host"`
	RedisPassword string `mapstructure:"password"`
	RedisDB       int    `mapstructure:"db"`
}

type DatabaseProvider struct {
	DatabaseAddress  string `mapstructure:"host"`
	DatabaseUsername string `mapstructure:"username"`
	DatabasePassword string `mapstructure:"password"`
	DatabaseName     string `mapstructure:"database"`
	DatabaseSSL      bool   `mapstructure:"ssl"`
	/* DatabaseColumnID       string
	DatabaseColumnFilename string */
}

type HTTPServer struct {
	DeliveryPort   int    `mapstructure:"port"`
	ServerSubPath  string `mapstructure:"path"`
	AllowDeletion  bool   `mapstructure:"allow_delete"`
	AllowInsertion bool   `mapstructure:"allow_insert"`
}

type Telemetry struct {
	JaegerAddress string `mapstructure:"jaeger_address"`
	JaegerPort    int    `mapstructure:"jaeger_port"`
}

func NewConfig() (Config, error) {
	consul_service_id := utils.RandStringBytes(4)
	cfg := Config{
		Consul{
			ConsulServiceID: consul_service_id,
		},
		RedisDatabase{RedisEnable: false},
		DatabaseProvider{DatabaseSSL: false},
		HTTPServer{DeliveryPort: 3000},
		Telemetry{JaegerPort: 4318},
	}

	err := cfg.loadFromFile()
	if cfg.Consul.ConsulServiceAddress == "auto" {
		cfg.Consul.ConsulServiceAddress = utils.GetLocalIPv4()
	}

	return cfg, err
}

func (cfg *Config) loadFromFile() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config/")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("APP") // Allow override from environemnt via APP_VAR_NAME

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	err := viper.Unmarshal(&cfg)
	return err
}

func (cfg *Config) GetServiceDefinition() capi.AgentServiceRegistration {
	csl := cfg.Consul
	return capi.AgentServiceRegistration{
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
