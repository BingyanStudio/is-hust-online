package db

import (
	"context"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/config"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedisClient(conf config.RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	return err
}

func CloseRedisClient() error {
	return RedisClient.Close()
}
