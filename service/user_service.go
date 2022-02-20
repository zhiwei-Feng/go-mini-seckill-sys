package service

import (
	"context"
	"github.com/rs/zerolog/log"
	"mini-seckill/config"
	"mini-seckill/util"
	"strconv"
	"time"
)

func UserCountSelfIncrement(userId int) int {
	limitKey := LIMIT_KEY + "_" + strconv.Itoa(userId)
	limitNumStr, err1 := util.GetRedisStringVal(limitKey)
	limitNum, err2 := strconv.Atoi(limitNumStr)
	if err1 != nil || err2 != nil {
		err := util.SetRedisStringVal(limitKey, "1")
		if err != nil {
			return -1
		}
		return 1
	} else {
		err := util.SetRedisStringVal(limitKey, strconv.Itoa(limitNum+1))
		if err != nil {
			return -1
		}
		return limitNum + 1
	}
}

func GetUserIsBanned(userId int) bool {
	limitKey := LIMIT_KEY + "_" + strconv.Itoa(userId)
	limitNumStr, err1 := util.GetRedisStringVal(limitKey)
	limitNum, err2 := strconv.Atoi(limitNumStr)
	if err1 != nil || err2 != nil {
		return true
	}

	return limitNum > ALLOW_COUNT
}

func UserAccessLimitCheck(userId int) (bool, error) {
	key := config.GenerateAccessLimitKey(userId)
	cli := util.RedisCli
	c := context.Background()
	// 无论是否首次都自增1
	n, err := cli.Exists(c, key).Result()
	if err != nil {
		log.Warn().Err(err)
		return false, err
	}
	if n == 0 {
		//首次访问
		err = cli.Set(c, key, 1, time.Minute*10).Err()
		if err != nil {
			log.Warn().Err(err)
			return false, err
		}
		return true, nil
	}
	val, err := cli.Incr(c, key).Result()
	if err != nil {
		log.Warn().Err(err)
		return false, err
	}
	return val > config.ACCESS_MAXCOUNT, nil
}
