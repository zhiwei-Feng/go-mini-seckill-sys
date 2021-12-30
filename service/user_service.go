package service

import (
	"log"
	"mini-seckill/util"
	"strconv"
)

func UserCountSelfIncrement(userId int) int {
	limitKey := LIMIT_KEY + "_" + strconv.Itoa(userId)
	limitNumStr, err1 := util.GetRedisStringVal(limitKey)
	limitNum, err2 := strconv.Atoi(limitNumStr)
	if err1 != nil || err2 != nil {
		log.Println(err1.Error(), err2.Error())
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
		log.Println("用户记录异常")
		return true
	}

	return limitNum > ALLOW_COUNT
}
