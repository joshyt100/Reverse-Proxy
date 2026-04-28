package config

import (
	"os"

	"go.yaml.in/yaml/v4"
)

type TLSConfig struct {
	Enabled    bool   `yaml:"enabled"`
	ListenAddr string `yaml:"listen_addr"`
	CertFile   string `yaml:"cert"`
	KeyFile    string `yaml:"key"`
}

// add Metrics
type MetricsConfig struct {
	Enabled    bool   `yaml:"enabled"`
	ListenAddr string `yaml:"listen_addr"`
}

type CleartextConfig struct {
	Enabled    bool   `yaml:"enabled"`
	ListenAddr string `yaml:"listen_addr"`
}

type RateLimitConfig struct {
	Enabled bool    `yaml:"enabled"`
	RPS     float64 `yaml:"rps"`   // request per second
	Burst   int     `yaml:"burst"` // max burst size

	PerIP bool `yaml:"per_ip"` // true -> per client IP, false -> global
}

type LoggingConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn error
	Format string `yaml:"format"` // json, text
}

type HealthConfig struct {
	Enabled             bool   `yaml:"enabled"`
	Path                string `yaml:"path"`
	IntervalSeconds     int    `yaml:"interval_seconds"`
	TimeoutSeconds      int    `yaml:"timeout_seconds"`
	PassiveCooldownSecs int    `yaml:"passive_cooldown_seconds"`
}

type Config struct {
	Cleartext  CleartextConfig `yaml:"cleartext"`
	ListenAddr string          `yaml:"listen_addr"`
	Upstreams  []string        `yaml:"upstreams"`
	Algo       string          `yaml:"algo"`
	TLS        TLSConfig       `yaml:"tls"`
	Metrics    MetricsConfig   `yaml:"metrics"`
	RateLimit  RateLimitConfig `yaml:"rate_limit"`
	Logger     LoggingConfig   `yaml:"logging"`
	Health     HealthConfig    `yaml:"health"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Cleartext: CleartextConfig{
			Enabled:    true,
			ListenAddr: ":8080",
		},
		Algo: "lc",
		TLS: TLSConfig{
			ListenAddr: ":8443",
		},
		Metrics: MetricsConfig{
			Enabled:    true,
			ListenAddr: ":2112",
		},
		Logger: LoggingConfig{Level: "info", Format: "text"},

		RateLimit: RateLimitConfig{Enabled: false},

		Health: HealthConfig{
			Enabled:             false,
			Path:                "/",
			IntervalSeconds:     5,
			TimeoutSeconds:      2,
			PassiveCooldownSecs: 10,
		},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
