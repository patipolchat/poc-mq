package util

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	DB    *DBConfig    `mapstructure:"db" validate:"required"`
	Kafka *KafkaConfig `mapstructure:"kafka" validate:"required"`
	Nats  *NatConfig   `mapstructure:"nats"`
}

type DBConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required"`
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	Dbname   string `mapstructure:"dbname" validate:"required"`
}

type KafkaConfig struct {
	Brokers   []string `mapstructure:"brokers" validate:"required"`
	Topic     string   `mapstructure:"topic" validate:"required"`
	Partition int      `mapstructure:"partition" validate:"required"`
}

type NatConfig struct {
	Url     string `mapstructure:"url" validate:"required"`
	Subject string `mapstructure:"subject" validate:"required"`
}

func NewConfig() (*Config, error) {
	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, err
	}

	return &config, nil
}
