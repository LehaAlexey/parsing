package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	HTTP     HTTPConfig     `yaml:"http"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
}

type KafkaConfig struct {
	Host               string `yaml:"host"`
	Port               int    `yaml:"port"`
	ParseRequestedTopic string `yaml:"parse_requested_topic_name"`
}

type HTTPConfig struct {
	Addr string `yaml:"addr"`
}

type GRPCConfig struct {
	Addr string `yaml:"addr"`
}

type SchedulerConfig struct {
	TickSeconds            int `yaml:"tick_seconds"`
	DefaultIntervalSeconds int `yaml:"default_interval_seconds"`
	MaxBatch               int `yaml:"max_batch"`
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
