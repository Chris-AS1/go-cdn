package config

import (
	"encoding/json"
	"fmt"
	"go-cdn/utils"
	"log"

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
	ConsulDB          string `mapstructure:"db"`
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
	dbg, _ := json.MarshalIndent(cfg, "", "  ")
	log.Print("Loaded following configs:", string(dbg))
	return cfg, err
}

func (cfg *Config) loadFromFile() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config/")
	viper.AddConfigPath("../config/")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")

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
		Address: csl.ConsulAddress,
		Port:    csl.ConsulPort,
	}
}

func (cfg *Config) GetConsulConfig() *capi.Config {
	return &capi.Config{
		Address:    fmt.Sprintf("%s:%d", cfg.Consul.ConsulAddress, cfg.Consul.ConsulPort),
		Datacenter: cfg.Consul.ConsulDB,
		Scheme:     "http",
	}
}
