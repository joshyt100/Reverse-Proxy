package config

import (
	"go.yaml.in/yaml/v4"
	"os"
)

type Config struct {
	ListenAddr string   `yaml:"ListenAddr"`
	Upstreams  []string `yaml:"Upstreams"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		ListenAddr: ":8080", // default
		Upstreams: []string{
			"http://localhost:9000",
			"http://localhost:9001",
			"http://localhost:9002",
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
