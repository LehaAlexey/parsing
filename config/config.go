package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Kafka  KafkaConfig  `yaml:"kafka"`
	HTTP   HTTPConfig   `yaml:"http"`
	Parser ParserConfig `yaml:"parser"`
	Swagger SwaggerConfig `yaml:"swagger"`
}

type KafkaConfig struct {
	Host                   string `yaml:"host"`
	Port                   int    `yaml:"port"`
	ParseRequestedTopic    string `yaml:"parse_requested_topic_name"`
	PriceMeasuredTopic     string `yaml:"price_measured_topic_name"`
	GroupID                string `yaml:"group_id"`
}

type HTTPConfig struct {
	Addr string `yaml:"addr"`
}

type ParserConfig struct {
	UserAgent              string `yaml:"user_agent"`
	RequestTimeoutMS       int    `yaml:"request_timeout_ms"`
	MaxBodyBytes           int64  `yaml:"max_body_bytes"`
	Retries                int    `yaml:"retries"`
	MinBackoffMS           int    `yaml:"min_backoff_ms"`
	MaxBackoffMS           int    `yaml:"max_backoff_ms"`
	PerDomainMinIntervalMS int    `yaml:"per_domain_min_interval_ms"`
}

type SwaggerConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &cfg, nil
}
