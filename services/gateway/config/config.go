package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	HTTP       HTTPConfig       `yaml:"http"`
	UsersGRPC  GRPCConfig       `yaml:"users_grpc"`
	HistoryGRPC GRPCConfig      `yaml:"history_grpc"`
}

type HTTPConfig struct {
	Addr string `yaml:"addr"`
}

type GRPCConfig struct {
	Addr string `yaml:"addr"`
}

func LoadConfig(filename string) (*Config, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &cfg, nil
}

