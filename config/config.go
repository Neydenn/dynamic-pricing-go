package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type Postgres struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
}

type KafkaCatalog struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

type KafkaOrder struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

type KafkaPricing struct {
	Brokers      []string `yaml:"brokers"`
	OrdersTopic  string   `yaml:"orders_topic"`
	CatalogTopic string   `yaml:"catalog_topic"`
	PricingTopic string   `yaml:"pricing_topic"`
	GroupID      string   `yaml:"group_id"`
}

type Catalog struct {
	HTTPAddr string       `yaml:"http_addr"`
	DB       Postgres     `yaml:"db"`
	Kafka    KafkaCatalog `yaml:"kafka"`
}

type Order struct {
	HTTPAddr string     `yaml:"http_addr"`
	DB       Postgres   `yaml:"db"`
	Kafka    KafkaOrder `yaml:"kafka"`
}

type Pricing struct {
	HTTPAddr string       `yaml:"http_addr"`
	DB       Postgres     `yaml:"db"`
	Kafka    KafkaPricing `yaml:"kafka"`
}

type Root struct {
	Catalog Catalog `yaml:"catalog"`
	Order   Order   `yaml:"order"`
	Pricing Pricing `yaml:"pricing"`
}

func Load(path string) (Root, error) {
	var cfg Root
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	if cfg.Catalog.HTTPAddr == "" || cfg.Order.HTTPAddr == "" || cfg.Pricing.HTTPAddr == "" {
		return cfg, errors.New("invalid config")
	}
	return cfg, nil
}
