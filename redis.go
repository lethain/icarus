package icarus

import (
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"

	"log"
)

var redisProto = "tcp"
var redisLocation = "localhost:6379"
var poolSize = 10
var redisPool *pool.Pool

func ConfigRedis(cfg *Config) error {
	if cfg.Redis.Loc != "" {
		redisLocation = cfg.Redis.Loc
	}
	if cfg.Redis.Proto != "" {
		redisProto = cfg.Redis.Proto
	}
	if cfg.Redis.PoolSize != 0 {
		poolSize = cfg.Redis.PoolSize
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
