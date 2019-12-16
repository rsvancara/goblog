package cache

import (
	"github.com/gomodule/redigo/redis"
)

//GetRedisConn get redis connection
func GetRedisConn() (redis.Conn, error) {
	// Establish a connection to the Redis server listening on port
	// 6379 of the local machine. 6379 is the default port, so unless
	// you've already changed the Redis configuration file this should
	// work.
	conn, err := redis.Dial("tcp", "host.docker.internal:6379")
	if err != nil {
		return conn, err
	}
	// Importantly, use defer to ensure the connection is always
	// properly closed before exiting the main() function.

	return conn, nil
}
