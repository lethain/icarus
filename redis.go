package icarus

import (
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/mediocregopher/radix.v2/pool"

	"log"
)

var redisProto = "tcp"
var redisLocation = "127.0.0.1:6379"
var poolSize = 10
var redisPool *pool.Pool

func ConfigRedis(cfg *Config) error {
	if cfg.RedisLoc != "" {
		redisLocation = cfg.RedisLoc
	}
	var err error
	redisPool, err = pool.New(redisProto, redisLocation, poolSize)
	return err
}

func GetRedisClient() (*redis.Client, error) {
	if redisPool == nil {
		log.Fatal("must ConfigRedis before retrieving clients")
	}
	return redisPool.Get()
}

func PutRedisClient(rc *redis.Client) {
	if redisPool == nil {
		log.Fatal("must ConfigRedis before returning clients")
	}	
	redisPool.Put(rc)
}
