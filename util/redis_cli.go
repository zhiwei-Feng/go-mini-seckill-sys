package util

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func GetRedisStringVal(key string) (string, error) {
	val, err := rdb.Get(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func SetRedisStringVal(key string, val string) error {
	err := rdb.Set(context.Background(), key, val, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func DelRedisKey(key string) error {
	err := rdb.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}
