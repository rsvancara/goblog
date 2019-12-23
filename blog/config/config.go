package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

//AppConfig Application Configuration
type AppConfig struct {
	Cacheuri string `envconfig:"CACHE_URI"`
	Dburi    string `envconfig:"DB_URI"`
}

// GetConfig get the current configuration from the environment
func GetConfig() (AppConfig, error) {
	var cfg AppConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		fmt.Println(err)
	}

	return cfg, nil
}
