package config

import "github.com/spf13/viper"

type Config struct {
	Env string `yaml:"env" env:"ENV"`
}

func LoadConfig() *Config {
	viper.SetConfigFile("configs.yaml")
	viper.
	return &Config{

	}
}
