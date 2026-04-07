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

	if err := l.Viper.Unmarshal(l.ConfigService); err != nil {
		return nil, fmt.Errorf("unable to decode into configService, %w", err)
	}

	return l, nil
}

func (l *Loader) loadYamlFiles(env string) {
	yamlFiles := []string{
		filepath.Join(".", "..", "..", "shared", "internal", "configs", env),
		filepath.Join(".", env),
	}

	for _, yamlFile := range yamlFiles {
		l.Viper.SetConfigName(yamlFile)
		l.Viper.SetConfigType("yaml")
		if err := l.Viper.MergeInConfig(); err != nil {
			log.Printf("loadYamlFiles, MergeInConfig is failed, yamlFile path: %s, err: %v", yamlFile, err)
		}
	}
}

func (l *Loader) loadEnvFiles(env string) {
	envFiles := []string{
		filepath.Join("..", "..", "shared", "internal", "configs", env),
		filepath.Join(".", env),
	}

	for _, envFile := range envFiles {
		l.Viper.SetConfigName(envFile)
		l.Viper.SetConfigType("env")
		if err := l.Viper.MergeInConfig(); err != nil {
			log.Printf("loadEnvFiles, MergeInConfig is failed, err: %v", err)
		}
	}
}

func (l *Loader) setupSettings() {
	l.Viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) //TODO: зачем нужен реплэйсер? просто чтобы в env файле не было ключей с точкой? и второй вопрос здесь же - где использовать strings.ToLower?
	l.Viper.SetEnvPrefix(l.ServiceName)
}
