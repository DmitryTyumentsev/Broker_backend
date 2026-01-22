package config

import (
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

	addYaml(l)

	addEnv()

	if err := viper.Unmarshal(&l.ConfigService); err != nil {
		return nil, err
	}

	return l, nil
}

func addYaml(l *Loader) {
	yamlFiles := []string{
		filepath.Join("..", "..", "shared", "internal", "config", "local.yaml"),
		filepath.Join("..", "..", "shared", "internal", "config", "dev.yaml"),
		filepath.Join(".", "local.yaml"),
		filepath.Join(".", "dev.yaml"),
		filepath.Join(".", "prod.yaml"), //TODO: 99% что убрать эту строчку, yaml не пишут же у прода?
	}

	for _, yamlFile := range yamlFiles {
		_, err := os.Stat(yamlFile)
		if os.IsNotExist(err) {
			log.Println("Yaml file not found in directory:", yamlFile)
			continue
		}
		l.Viper.SetConfigFile(yamlFile)

	}
}
