package cache

import (
	"fmt"

	"goblog/internal/config"

	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"
)

type Cache struct {
	Pool *redis.Pool
	cfg  config.AppConfig
}

func (cache *Cache) GetRedisConn() (*redis.Conn, error) {
	return nil, nil
}

func (cache *Cache) InitPool(cfg config.AppConfig) error {

	pool := &redis.Pool{
		MaxIdle:   6,
		MaxActive: 100,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", cfg.Cacheuri)
			if err != nil {
				log.Error().Err(err).Msgf("ERROR: fail init redis pool: %s", err.Error())
				return nil, err
			}
			return conn, err
		},
	}

	cache.Pool = pool
	cache.cfg = cfg

	return nil
}

func (cache *Cache) getConn() (redis.Conn, error) {
	conn := cache.Pool.Get()
	//defer conn.Close()

	_, err := redis.String(conn.Do("PING"))
	if err != nil {
		log.Error().Err(err).Msgf("fail ping check redis conn to cache uri %s", cache.cfg.Cacheuri)
		return nil, err
	}

	conn.Do("SELECT", cache.cfg.RedisDB)

	return conn, nil
}

// Internally destroy the redis connection and capture any errors
func (cache *Cache) cleanup(conn redis.Conn) {

	//log.Debug().Msg("cleaning up redis connection")
	err := conn.Close()
	if err != nil {
		log.Error().Err(err).Msgf("error closing redis connection")
	}
}

func (cache *Cache) GetTTL(key string) (int, error) {
	// Connect to Redis using connection pool
	conn, err := cache.getConn()
	if err != nil {
		return 0, fmt.Errorf("redis is not available or does not respond to ping: %s", err)
	}
	defer cache.cleanup(conn)

	ttl, err := redis.Int(conn.Do("TTL", key))
	if err != nil {
		return 0, fmt.Errorf("error unmarshaling session %s with error %s", key, err)
	}

	return ttl, nil
}

// Ping the redis cach to see if the connection is alive
func (cache *Cache) Ping() error {
	// Connect to Redis using connection pool
	conn, err := cache.getConn()
	if err != nil {
		return fmt.Errorf("redis is not available or does not respond to ping: %s", err)
	}
	defer cache.cleanup(conn)

	// Ping cache to ensure pool is working
	err = cache.Ping()
	if err != nil {
		return fmt.Errorf("redis is not available or does not respond to ping: %s", err)
	}

	return nil
}

func (cache *Cache) GetKey(key string) (string, error) {

	// Connect to Redis using connection pool
	conn, err := cache.getConn()
	if err != nil {
		return "", fmt.Errorf("redis is not available or does not respond to ping: %s", err)
	}
	defer cache.cleanup(conn)

	response, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return "", fmt.Errorf("error retrieving user object from redis: %s", err)
	}

	return response, nil
}

// Set Key Value method for setting a key
func (cache *Cache) SetEx(key string, value string, timeout int64) error {

	// Connect to Redis using connection pool
	conn, err := cache.getConn()
	if err != nil {
		return fmt.Errorf("redis is not available or does not respond to ping: %s", err)
	}
	defer cache.cleanup(conn)

	_, err = conn.Do("SETEX", key, timeout, value)
	if err != nil {
		return fmt.Errorf("error saving session to redis: %s", err)
	}
	return nil
}

// Redis Set.  Set does not change the TTL.  Useful for when you want
// to update the cache item but dont want to update the TTL
func (cache *Cache) Set(key string, value string) error {
	// Connect to Redis using connection pool
	conn, err := cache.getConn()
	if err != nil {
		return fmt.Errorf("redis is not available or does not respond to ping: %s", err)
	}
	defer cache.cleanup(conn)

	_, err = conn.Do("SET", key, value)
	if err != nil {
		return fmt.Errorf("error saving session to redis: %s", err)
	}
	return nil
}

// GetKey internal method for getting keys for a supplied pattern, like "*"
func (cache *Cache) GetAllVals(pattern string) (map[string]string, error) {

	// Connect to Redis using connection pool
	conn, err := cache.getConn()
	if err != nil {
		return nil, fmt.Errorf("redis is not available or does not respond to ping: %s", err)
	}
	defer cache.cleanup(conn)

	iter := 0

	m := make(map[string]string)
	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return nil, fmt.Errorf("could not retrieve keys for pattern '%s' with erorr %s", pattern, err.Error())
		}

		iter, _ = redis.Int(arr[0], nil)
		keys, _ := redis.Strings(arr[1], nil)

		var args []interface{}
		// Load keys into interface
		for _, k := range keys {
			args = append(args, k)
		}

		// Pass into mget for more efficient retrieval
		// according to documentation order is preserved with missing keys returning nil
		vals, _ := redis.Strings(conn.Do("MGET", args...))

		for i, k := range keys {
			m[k] = vals[i]
		}

		if iter == 0 {
			break
		}
	}

	return m, nil
}

func (cache *Cache) Exists(key string) (bool, error) {

	// Connect to Redis using connection pool
	conn, err := cache.getConn()
	if err != nil {
		return false, fmt.Errorf("redis is not available or does not respond to ping: %s", err)
	}
	defer cache.cleanup(conn)

	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return ok, fmt.Errorf("error checking if key %s exists: %v", key, err)
	}
	return ok, err
}

func (cache *Cache) Delete(key string) error {

	// Connect to Redis using connection pool
	conn, err := cache.getConn()
	if err != nil {
		return fmt.Errorf("redis is not available or does not respond to ping: %s", err)
	}
	defer cache.cleanup(conn)

	_, err = conn.Do("DEL", key)
	if err != nil {
		return fmt.Errorf("error deleting key %s with error: %s", key, err.Error())
	}
	return err
}

func (cache *Cache) GetPoolStatus() (redis.PoolStats, error) {

	return cache.Pool.Stats(), nil

}
