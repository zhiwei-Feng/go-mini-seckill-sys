package util

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

func newRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return rdb
}

func GetRedisStringVal(key string) (string, error) {
	rdb := newRedisClient()
	val, err := rdb.Get(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func SetRedisStringVal(key string, val string) error {
	rdb := newRedisClient()
	err := rdb.Set(context.Background(), key, val, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}
