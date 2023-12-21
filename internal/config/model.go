package config

type Config struct {
	Consul     Consul     `mapstructure:"consul"`
	Cache      Cache      `mapstructure:"redis"`
	Database   Database   `mapstructure:"postgres"`
	HTTPServer HTTPServer `mapstructure:"http"`
	Telemetry  Telemetry  `mapstructure:"telemetry"`
}

type Consul struct {
	ConsulEnable         bool   `mapstructure:"enable"`
	ConsulServiceID      string `mapstructure:"service_id"` // Should not be present in configs.yaml. Randomized for each instance.
	ConsulServiceName    string `mapstructure:"service_name"`
	ConsulServiceAddress string `mapstructure:"service_address"`
	ConsulAddress        string `mapstructure:"address"`
	ConsulDatacenter     string `mapstructure:"datacenter"`
	ConsulPort           int    `mapstructure:"port"`
}

type Cache struct {
	RedisEnable   bool   `mapstructure:"enable"`
	RedisAddress  string `mapstructure:"host"`
	RedisPassword string `mapstructure:"password"`
	RedisDB       int    `mapstructure:"db"`
}

type Database struct {
	DatabaseAddress  string `mapstructure:"host"`
	DatabaseUsername string `mapstructure:"username"`
	DatabasePassword string `mapstructure:"password"`
	DatabaseName     string `mapstructure:"database"`
	DatabaseSSL      bool   `mapstructure:"ssl"`
}

type HTTPServer struct {
	DeliveryPort    int    `mapstructure:"port"`
	ServerSubPath   string `mapstructure:"path"`
	AllowDeletion   bool   `mapstructure:"allow_delete"`
	AllowInsertion  bool   `mapstructure:"allow_insert"`
	RateLimitEnable bool   `mapstructure:"rate_limit_enable"`
	RateLimit       int    `mapstructure:"rate_limit"`
}

type Telemetry struct {
	TelemetryEnable bool    `mapstructure:"enable"`
	JaegerAddress   string  `mapstructure:"jaeger_address"`
	Sampling        float64 `mapstructure:"sampling"`
	LogPath         string  `mapstructure:"logs_path"`
	LogMaxSize      int     `mapstructure:"logs_max_size"`
	LogMaxBackups   int     `mapstructure:"logs_max_backups"`
	LogMaxAge       int     `mapstructure:"logs_max_age"`
}

