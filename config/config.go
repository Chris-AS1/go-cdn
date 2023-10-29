package config

import (
	"fmt"
	"go-cdn/utils"

	capi "github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
)

type Config struct {
	Consul           Consul           `mapstructure:"consul"`
	Redis            RedisDatabase    `mapstructure:"redis"`
	DatabaseProvider DatabaseProvider `mapstructure:"postgres"`
	HTTPServer       HTTPServer       `mapstructure:"http"`
}

type Consul struct {
	ConsulEnable      bool   `mapstructure:"enable"`
	ConsulServiceID   string `mapstructure:"service_id"` // Should not be present in configs.yaml. It's randomized for each instance
	ConsulServiceName string `mapstructure:"service_name"`
	ConsulAddress     string `mapstructure:"address"`
	ConsulDatacenter  string `mapstructure:"datacenter"`
	ConsulPort        int    `mapstructure:"port"`
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

func NewConfig() (Config, error) {
	consul_service_id := utils.RandStringBytes(4)
	cfg := Config{
		Consul{
			ConsulServiceID: consul_service_id,
		},
		RedisDatabase{RedisEnable: false},
		DatabaseProvider{DatabaseSSL: false},
		HTTPServer{DeliveryPort: 3000},
	}

	err := cfg.loadFromFile()
	return cfg, err
}

func (cfg *Config) loadFromFile() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config/")
	viper.AddConfigPath("../config/") // To remove eventually
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	err := viper.Unmarshal(&cfg)
	return err
}

func (cfg *Config) GetServiceDefinition() capi.AgentServiceRegistration {
	this_addr := utils.GetLocalIPv4() // Find the local address if deployed in docker
	csl := cfg.Consul
	return capi.AgentServiceRegistration{
		ID:      csl.ConsulServiceID,
		Name:    csl.ConsulServiceName,
		Address: this_addr,
		Port:    cfg.HTTPServer.DeliveryPort,
		Check: &capi.AgentServiceCheck{
			Name:                           "web_alive",
			Interval:                       "10s",
			Timeout:                        "30s",
			HTTP:                           fmt.Sprintf("http://%s:%d/health", this_addr, cfg.HTTPServer.DeliveryPort),
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
