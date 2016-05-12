package icarus

import (
	"github.com/mediocregopher/radix.v2/redis"
)

const RedisNetwork = "tcp"

var RedisLocation = "127.0.0.1:6379"

func GetRedisClient(loc string) (*redis.Client, error) {
	return redis.Dial(RedisNetwork, loc)
}

func GetConfiguredRedisClient() (*redis.Client, error) {
	return GetRedisClient(RedisLocation)
}
