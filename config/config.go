package config

import "strconv"

// === 验证URL Hashcode ===
const URL_VERIFY_PREFIX = "seckill_hash"

// === 控制用户恶意多次访问 ===
const ACCESS_LIMIT_PREFIX = "seckill_limit"
const ACCESS_MAXCOUNT = 10

// === 防止用户重复抢购 ===
const HAS_ORDER_PREFIX = "has_order" // 重复抢购下单

// === 商品缓存 ===
const STOCK = "stock"

// === 用户订单创建锁 ===
const ORDER_CREATE = "order_create"

func GenerateURLVerifyKey(stockId, userId int) string {
	return URL_VERIFY_PREFIX + "_" + strconv.Itoa(stockId) + "_" + strconv.Itoa(userId)
}

func GenerateAccessLimitKey(userId int) string {
	return ACCESS_LIMIT_PREFIX + "_" + strconv.Itoa(userId)
}

func GenerateHasOrderKey(stockId int) string {
	return HAS_ORDER_PREFIX + "_" + strconv.Itoa(stockId)
}

func GenerateStockKey(stockId int) string {
	return STOCK + "_" + strconv.Itoa(stockId)
}

func GenerateOrderCreateKey(stockId, userId int) string {
	return ORDER_CREATE + "_" + strconv.Itoa(stockId) + "_" + strconv.Itoa(userId)
}

// === 消息队列名称 ===
const OrderCreateQueueName = "orderCreate" // 异步下单队列
