package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"
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

	l.loadYamlFiles()

	//loadEnvFiles()

	if err := viper.Unmarshal(&l.ConfigService); err != nil {
		return nil, err
	}

	return l, nil
}

func(l *Loader) loadYamlFiles() {
	yamlFiles := []string{
		filepath.Join("..", "..", "shared", "internal", "config", "local.yaml"),
		filepath.Join("..", "..", "shared", "internal", "config", "dev.yaml"),
		filepath.Join(".", "local.yaml"),
		filepath.Join(".", "dev.yaml"),
	}

	for _, yamlFile := range yamlFiles {
		_, err := os.Stat(yamlFile)
		if errors.Is()
		if err != nil{
			if os.IsNotExist(err) {
				log.Println("Yaml file not found in directory:", yamlFile)
				continue
			}

		}

		l.Viper.SetConfigFile(yamlFile)

	}
}
