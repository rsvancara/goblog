package blog

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

func GetRedisConn() redis.Conn {
	// Establish a connection to the Redis server listening on port
	// 6379 of the local machine. 6379 is the default port, so unless
	// you've already changed the Redis configuration file this should
	// work.
	conn, err := redis.Dial("tcp", "10.152.64.116:32777")
	if err != nil {
		fmt.Println(err)
	}
	// Importantly, use defer to ensure the connection is always
	// properly closed before exiting the main() function.

	return conn
}
