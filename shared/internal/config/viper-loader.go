package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

type Loader struct {
	ConfigService interface{}
	ServiceName   string
	Viper         *viper.Viper
	Mu            sync.Mutex
	Watchers      []func(cfg interface{}) error
}

func NewViperLoader(cfgPtr interface{}, serviceName string) (*Loader, error) {
	l := &Loader{
		ConfigService: cfgPtr,
		ServiceName:   serviceName,
		Viper:         viper.New(),
	}

	l.setupSettings()

	env := os.Getenv("ENVIRONMENT")

	l.loadYamlFiles(env)

	l.loadEnvFiles(env)

	l.Viper.AutomaticEnv()

	if err := viper.Unmarshal(l.ConfigService); err != nil {
		return nil, fmt.Errorf("unable to decode into configService, %w", err)
	}

	return l, nil
}

func (l *Loader) loadYamlFiles(env string) {
	yamlFiles := []string{
		filepath.Join("..", "..", "shared", "internal", "config", env),
		filepath.Join(".", env),
	}

	v := l.Viper
	for _, yamlFile := range yamlFiles {
		v.SetConfigName(yamlFile)
		v.SetConfigType("yaml")
		if err := v.MergeInConfig(); err != nil {
			log.Printf("loadYamlFiles, MergeInConfig is failed, err: %v", err)
		}
	}
	v.AutomaticEnv()
}

func (l *Loader) loadEnvFiles(env string) {
	envFiles := []string{
		filepath.Join("..", "..", "shared", "internal", "config", env),
		filepath.Join(".", env),
	}

	v := l.Viper
	for _, envFile := range envFiles {
		v.SetConfigName(envFile)
		v.SetConfigType("env")
		if err := v.MergeInConfig(); err != nil {
			log.Printf("loadEnvFiles, MergeInConfig is failed, err: %v", err)
		}
	}
}

func (l *Loader) setupSettings() {
	l.Viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) //TODO: зачем нужен реплэйсер? просто чтобы в env файле не было ключей с точкой?
	l.Viper.SetEnvPrefix(l.ServiceName)
	l.Viper.RegisterAlias()
}
