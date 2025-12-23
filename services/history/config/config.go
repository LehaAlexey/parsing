package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Kafka    KafkaConfig    `yaml:"kafka"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	HTTP     HTTPConfig     `yaml:"http"`
}

type KafkaConfig struct {
	Host                  string `yaml:"host"`
	Port                  int    `yaml:"port"`
	PriceMeasuredTopic    string `yaml:"price_measured_topic_name"`
	GroupID               string `yaml:"group_id"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
}

type RedisConfig struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	DB                int    `yaml:"db"`
	RecentLimit       int    `yaml:"recent_limit"`
	RecentTTLSeconds  int    `yaml:"recent_ttl_seconds"`
}

type GRPCConfig struct {
	Addr string `yaml:"addr"`
}

type HTTPConfig struct {
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

