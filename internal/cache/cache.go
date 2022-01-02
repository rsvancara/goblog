package cache

import (
	"fmt"

	"goblog/internal/config"

	"github.com/gomodule/redigo/redis"
)

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
