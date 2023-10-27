package config

import (
	"encoding/json"
	"go-cdn/utils"
	"log"

	capi "github.com/hashicorp/consul/api"
	"github.com/spf13/viper"
)

type Config struct {
	Consul   Consul           `mapstructure:"consul"`
	Redis    RedisDatabase    `mapstructure:"redis"`
	Database DatabaseProvider `mapstructure:"database"`
}

type Consul struct {
	ConsulServiceID   string `mapstructure:"service_id"`
	ConsulServiceName string `mapstructure:"service_name"`
	ConsulAddress     string `mapstructure:"address"`
	ConsulDB          string `mapstructure:"db"`
	ConsulPort        int    `mapstructure:"port"`
}

type RedisDatabase struct {
	RedisAddress  string `mapstructure:"url"`
	RedisPassword string `mapstructure:"password"`
	RedisEnable   bool   `mapstructure:"enable"`
	RedisDB       int    `mapstructure:"db"`
}

type DatabaseProvider struct {
	DatabaseUsername string `mapstructure:"username"`
	DatabasePassword string `mapstructure:"password"`
	DatabasePort     string `mapstructure:"port"`
	DatabaseURN      string `mapstructure:"urn"`
	DatabaseSSL      string `mapstructure:"ssl"`
}

func NewConfig() Config {
	consul_service_id := utils.RandStringBytes(4)
	cfg := Config{
		Consul{
			ConsulServiceID: consul_service_id,
		},
		RedisDatabase{},
		DatabaseProvider{},
	}

	cfg.loadFromFile()
	js, _ := json.Marshal(cfg)
	log.Print("Loaded following configs:", string(js))
	return cfg
}

func (cfg *Config) loadFromFile() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config/")
	viper.AddConfigPath("../config/")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")

	if err := viper.ReadInConfig(); err != nil {
		log.Panic("Error reading config file, %s", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Panic("Error reading config file, %s", err)
	}
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
		Address:    cfg.Consul.ConsulAddress,
		Datacenter: cfg.Consul.ConsulDB,
	}
}
