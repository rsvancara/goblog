package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

//AppConfig Application Configuration
type AppConfig struct {
	Cacheuri      string `envconfig:"CACHE_URI"`      // Cacheuri
	Dburi         string `envconfig:"DB_URI"`         //MongDB URI
	AdminUser     string `envconfig:"ADMIN_USER"`     // Admin User for application
	AdminPassword string `envconfig:"ADMIN_PASSWORD"` // Admin Password for application
	S3Bucket      string `envconfig:"S3_URI"`         // Where your S3 Buckets is
	Env           string `envconfig:"ENV"`            //PROD,DEV
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
