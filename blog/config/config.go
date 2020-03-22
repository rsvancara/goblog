package config

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kelseyhightower/envconfig"
)

//AppConfig Application Configuration
type AppConfig struct {
	Cacheuri       string `envconfig:"CACHE_URI"`       // Cacheuri
	RedisPassword  string `envconfig:"REDIS_PASSWORD"`  // Redis Password
	Dburi          string `envconfig:"DB_URI"`          //MongDB URI
	AdminUser      string `envconfig:"ADMIN_USER"`      // Admin User for application
	AdminPassword  string `envconfig:"ADMIN_PASSWORD"`  // Admin Password for application
	S3Bucket       string `envconfig:"S3_BUCKET"`       // Where your S3 Buckets is
	Env            string `envconfig:"ENV"`             //PROD,DEV
	Site           string `envconfig:"SITE"`            // defines site name and location of template directories etc...
	SessionTimeout string `envconfig:"SESSION_TIMEOUT"` // defines session timeout
	MongoDatabase  string `envconfig:"MONGO_DATABASE"`  // Defines mongo database
	RedisDB        string `envconfig:"REDIS_DB"`        // Defines logical redis database
	RedisFilterDB  string `envconfig:"REDISFILTER_DB"`  // Defines logical redis filter database
}

//GetCacheURI returs cache uri for redis
func (a *AppConfig) GetCacheURI() string {
	return a.Cacheuri
}

//GetDBURI returns mongodb URI
func (a *AppConfig) GetDBURI() string {
	return a.GetDBURI()
}

//GetAdminUser returns admin user
func (a *AppConfig) GetAdminUser() string {
	return a.AdminUser
}

//GetAdminPassword returns admin password
func (a *AppConfig) GetAdminPassword() string {
	return a.AdminPassword
}

//GetS3Bucket returns s3 bucket where images are stored
func (a *AppConfig) GetS3Bucket() string {
	return a.S3Bucket
}

//GetEnv get the run time environment
func (a *AppConfig) GetEnv() string {
	return a.Env
}

//GetSite get the site to use for the template directory
func (a *AppConfig) GetSite() string {
	return a.Site
}

//GetSessionTimeout sets the session lifetime for redis and cookies
func (a *AppConfig) GetSessionTimeout() string {
	return a.SessionTimeout
}

//GetDurationTimeout sets the session lifetime for redis and cookies
func (a *AppConfig) GetDurationTimeout() time.Duration {

	//fmt.Println(a.SessionTimeout)
	//retVal, err := time.ParseDuration(a.SessionTimeout + "s")
	val, err := strconv.ParseUint(a.SessionTimeout, 10, 64)
	retVal := time.Duration(val)
	if err != nil {
		fmt.Printf("error parsing timout duration: %s\n", err)
		return 86400 * time.Second
	}

	//fmt.Println(retVal)
	return retVal * time.Second
}

//GetIntegerSessionTimeout sets the session lifetime for redis and cookies
func (a *AppConfig) GetIntegerSessionTimeout() int64 {
	retVal, err := strconv.ParseInt(a.SessionTimeout, 10, 32)
	if err != nil {
		fmt.Printf("error converting session to integer %s\n", err)
		return 86400

	}

	return retVal
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
