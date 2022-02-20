package config

import "strconv"

// 存储相关常量
const URL_VERIFY_PREFIX = "seckill_hash"

const ACCESS_LIMIT_PREFIX = "seckill_limit"
const ACCESS_MAXCOUNT = 10

const HAS_ORDER_PREFIX = "has_order" // 重复抢购下单

func GenerateURLVerifyKey(stockId, userId int) string {
	return URL_VERIFY_PREFIX + "_" + strconv.Itoa(stockId) + "_" + strconv.Itoa(userId)
}

func GenerateAccessLimitKey(userId int) string {
	return ACCESS_LIMIT_PREFIX + "_" + strconv.Itoa(userId)
}

func GenerateHasOrderKey(stockId int) string {
	return HAS_ORDER_PREFIX + "_" + strconv.Itoa(stockId)
}
