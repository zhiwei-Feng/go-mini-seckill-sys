package util

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

var RedisCli = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func GetRedisStringVal(key string) (string, error) {
	val, err := RedisCli.Get(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func SetRedisStringVal(key string, val string) error {
	err := RedisCli.Set(context.Background(), key, val, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func DelRedisKey(key string) error {
	err := RedisCli.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

func IsMember(key string, member string) (bool, error) {
	res, err := RedisCli.SIsMember(context.Background(), key, member).Result()
	if err != nil {
		return false, err
	}
	return res, nil
}

func SetAdd(key string, val string) error {
	err := RedisCli.SAdd(context.Background(), key, val).Err()
	if err != nil {
		return err
	}
	err = RedisCli.Expire(context.Background(), key, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}
