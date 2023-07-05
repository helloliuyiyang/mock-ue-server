package database

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"

	"mock-ue-server/pkg/logutil"
)

var logger = logutil.GetLogger()

var cli *redis.Client

func InitRedisCli(addr string) error {
	redisCli := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := redisCli.Ping(ctx).Result()
	if err != nil {
		return errors.Wrap(err, "redis ping failed")
	}
	cli = redisCli
	return nil
}

func GetRedisCli() *redis.Client {
	if cli == nil {
		logger.Panic("Redis client not init")
	}
	return cli
}
