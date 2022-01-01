package cache

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goblog/internal/config"

	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"
)

type CachePool struct {
	Pool *redis.Pool
}

//GetRedisConn get redis connection
func GetRedisConn() (redis.Conn, error) {
	// Establish a connection to the Redis server listening on port
	// 6379 of the local machine. 6379 is the default port, so unless
	// you've already changed the Redis configuration file this should
	// work.
	cfg, err := config.GetConfig()

	//conn, err := redis.Dial("tcp", cfg.Cacheuri, redis.DialPassword(cfg.RedisPassword))
	conn, err := redis.Dial("tcp", cfg.Cacheuri)
	if err != nil {

		fmt.Printf("Error connecting to redis with error %s using URI %s", err, cfg.Cacheuri)
		return conn, err
	}
	// Importantly, use defer to ensure the connection is always
	// properly closed before exiting the main() function.

	// Set the logical database in Redis
	conn.Do("SELECT", cfg.RedisDB)

	return conn, nil
}

//GetRedisFilterConn get redis connection
func GetRedisFilterConn() (redis.Conn, error) {
	// Establish a connection to the Redis server listening on port
	// 6379 of the local machine. 6379 is the default port, so unless
	// you've already changed the Redis configuration file this should
	// work.
	cfg, err := config.GetConfig()

	conn, err := redis.Dial("tcp", cfg.Cacheuri, redis.DialPassword(cfg.RedisPassword))
	if err != nil {

		fmt.Printf("Error connecting to redis with error %s using URI %s", err, cfg.Cacheuri)
		return conn, err
	}
	// Importantly, use defer to ensure the connection is always
	// properly closed before exiting the main() function.

	// Set the logical database in Redis
	conn.Do("SELECT", cfg.RedisFilterDB)

	return conn, nil
}

func (cache *CachePool) Init() error {

	// Establish a connection to the Redis server listening on port
	// 6379 of the local machine. 6379 is the default port, so unless
	// you've already changed the Redis configuration file this should
	// work.
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error().Err(err).Msg("error getting configuration")
	}

	cache.Pool.MaxIdle = 3
	cache.Pool.IdleTimeout = 240 * time.Second
	cache.Pool.Dial = func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", cfg.Cacheuri)
		if err != nil {
			return nil, err
		}
		return c, err
	}

	cache.Pool.TestOnBorrow = func(c redis.Conn, t time.Time) error {
		_, err := c.Do("PING")
		return err
	}

	return nil
}

func (cache *CachePool) CleanupHook() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		<-c
		cache.Pool.Close()
		os.Exit(0)
	}()
}
