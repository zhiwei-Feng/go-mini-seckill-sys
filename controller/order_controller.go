package controller

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"mini-seckill/config"
	"mini-seckill/domain"
	"mini-seckill/message"
	"mini-seckill/service"
	"mini-seckill/util"
	"mini-seckill/view"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// GoodsSeckillV2 商品抢购接口
// 功能流程：
// 1. 获取抢购商品的ID和当前抢购用户ID
// 2. 单机限流
// 3. 验证url hash
// 4. 单用户访问频率限制
// 5. 重复抢购检查
// 6. 检查库存
// 7. 异步下单（减库存）
func GoodsSeckillV2(c *gin.Context) {
	// 1.
	var param view.SeckillReq
	err := c.Bind(&param)
	if c.ShouldBind(&param) != nil || param.VerifyHash == "" {
		log.Warn().Err(err).Send()
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request params"})
		return
	} else {
		log.Info().Msgf("stockId|%d| userId|%d| |%s|",
			param.StockId, param.UserId, param.VerifyHash)
	}

	// 2.
	// rate limit
	if util.RateLimiter.Wait(c) != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "访问人数过多，请稍后重试"})
		return
	}

	// 3.
	hashKey := config.GenerateURLVerifyKey(param.StockId, param.UserId)
	verifiedHash, err := util.RedisCli.Get(c, hashKey).Result()
	if err != nil {
		log.Warn().Err(err).Msg("some errors in redis")
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if verifiedHash != param.VerifyHash {
		log.Warn().Msg("invalid request")
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	// 4.
	pass, err := service.UserAccessLimitCheck(param.UserId)
	if !pass {
		log.Info().Int("userId", param.UserId).Msg("请求次数异常过多！")
		c.JSON(http.StatusOK, gin.H{"message": "尝试过多，请30分钟后再试"})
		return
	}

	// 5.
	hasOrder, err := service.CheckOrderRepeat(param.StockId, param.UserId)
	if err != nil || hasOrder {
		c.JSON(http.StatusOK, gin.H{"message": "请不要重复抢购"})
		return
	}

	// 6.
	remain, err := service.GetStock(param.StockId)
	if err != nil || remain <= 0 {
		c.JSON(http.StatusOK, gin.H{"message": "当前库存不足，请稍后再试"})
		return
	}

	// 7.
	res, err := json.Marshal(domain.UserOrderInfo{Sid: param.StockId, UserId: param.UserId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	err = message.PublishMessage(res, config.OrderCreateQueueName)
	if err != nil {
		log.Error().Err(err).Msg("写入消息队列失败")
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "抢购进行中，等待订单生成"})
}

// GoodsSeckillV1 商品抢购接口
// 功能流程：
// 1. 获取抢购商品的ID和当前抢购用户ID
// 2. 单机限流
// 3. 验证url hash
// 4. 单用户访问频率限制
// 5. 重复抢购检查（高并发情况下可能会重复下单）
// 6. 下单（检查库存）+ 分布式锁防止重复下单（兜底）
func GoodsSeckillV1(c *gin.Context) {
	// 1.
	var param view.SeckillReq
	err := c.Bind(&param)
	if c.ShouldBind(&param) != nil || param.VerifyHash == "" {
		log.Warn().Err(err).Send()
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request params"})
		return
	} else {
		log.Info().Msgf("stockId|%d| userId|%d| |%s|",
			param.StockId, param.UserId, param.VerifyHash)
	}

	// 2.
	// rate limit
	if util.RateLimiter.Wait(c) != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "访问人数过多，请稍后重试"})
		return
	}

	// 3.
	hashKey := config.GenerateURLVerifyKey(param.StockId, param.UserId)
	verifiedHash, err := util.RedisCli.Get(c, hashKey).Result()
	if err != nil {
		log.Warn().Err(err).Msg("some errors in redis")
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if verifiedHash != param.VerifyHash {
		log.Warn().Msg("invalid request")
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	// 4.
	pass, err := service.UserAccessLimitCheck(param.UserId)
	if !pass {
		log.Info().Int("userId", param.UserId).Msg("请求次数异常过多！")
		c.JSON(http.StatusOK, gin.H{"message": "尝试过多，请30分钟后再试"})
		return
	}

	// 5.
	hasOrder, err := service.CheckOrderRepeat(param.StockId, param.UserId)
	if err != nil || hasOrder {
		c.JSON(http.StatusOK, gin.H{"message": "请不要重复抢购"})
		return
	}

	// 6.1 获取订单创建锁（用户+商品）
	lockKey := config.GenerateOrderCreateKey(param.StockId, param.UserId)
	lockSuccess, err := util.RedisCli.SetNX(c, lockKey, 1, time.Second*5).Result()
	if err != nil || !lockSuccess {
		c.JSON(http.StatusOK, gin.H{"message": "稍后再试"})
		return
	}

	// 6.2 创建订单
	remaining, err := service.CreateOrder(param.StockId, param.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "系统异常请重试"})
		return
	}

	// 6.3 释放锁
	delSuccess, err := util.RedisCli.Del(c, lockKey).Result()
	if err != nil || delSuccess == 0 {
		log.Error().Err(err).Msg("释放锁失败")
	}
	c.JSON(http.StatusOK, gin.H{"message": "抢购成功", "余下库存": remaining})
}

// CreateOrderWithMq 下单接口：异步订单创建
func CreateOrderWithMq(c *gin.Context) {
	sid, err := strconv.Atoi(c.Param("sid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid sid"})
		return
	}
	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid userId"})
		return
	}

	hasOrder, err := service.CheckOrderRepeat(sid, userId)
	if err != nil || hasOrder {
		c.JSON(http.StatusOK, gin.H{"message": "请不要重复抢购"})
		return
	}

	count := service.GetStockCountByCache(sid)
	if count == -1 {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	if count == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "无库存"})
		return
	}
	// 写入消息队列
	res, err := json.Marshal(domain.UserOrderInfo{sid, userId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	err = message.PublishMessage(res, "orderCreate")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "请求提交成功"})
}

// CreateOrderWithCacheV1
// 先删除缓存(库存)，再创建订单(写数据库)
func CreateOrderWithCacheV1(c *gin.Context) {
	sid, err := strconv.Atoi(c.Param("sid"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	res := service.DeleteStockCountCache(sid)
	if res {
		id := service.CreateOrderWithPessimisticLock(sid)
		if id == -1 {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "fail to create", "sid": sid})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "ok", "id": id})
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "fail to delete"})
	}
}

// CreateOrderWithCacheV2
// 先创建订单，再删除缓存(库存)
func CreateOrderWithCacheV2(c *gin.Context) {
	sid, err := strconv.Atoi(c.Param("sid"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}

	oid := service.CreateOrderWithPessimisticLock(sid)
	if oid == -1 {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "fail to create", "sid": sid})
	} else {
		res := service.DeleteStockCountCache(sid)
		if !res {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "fail to delete"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "ok", "id": oid})
		}
	}
}

// CreateOrderWithCacheV3
// 加入延时双删
func CreateOrderWithCacheV3(c *gin.Context) {
	sid, err := strconv.Atoi(c.Param("sid"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	res := service.DeleteStockCountCache(sid)
	if res {
		id := service.CreateOrderWithPessimisticLock(sid)
		if id == -1 {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "fail to create", "sid": sid})
		} else {
			// 延时再删除
			go func() {
				time.Sleep(time.Second)
				_ = service.DeleteStockCountCache(sid)

			}()
			c.JSON(http.StatusOK, gin.H{"message": "ok", "id": id})
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "fail to delete"})
	}
}

// CreateOrderWithCacheV4
// 加入删除缓存重试机制
func CreateOrderWithCacheV4(c *gin.Context) {
	sid, err := strconv.Atoi(c.Param("sid"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	service.DeleteStockCountCache(sid)
	id := service.CreateOrderWithPessimisticLock(sid)
	if id == -1 {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "fail to create", "sid": sid})
	} else {
		// 延时再删除
		go func() {
			time.Sleep(time.Second)
			//res := service.DeleteStockCountCache(sid)
			//if !res {
			//	log.Println("再删除失败，放入消息队列重试")
			//	err := util.PublishCacheDeleteMessage(strconv.Itoa(sid))
			//	if err != nil {
			//		log.Println("发布消息失败")
			//	}
			//} else {
			//	log.Println("再删除成功")
			//}
			_ = message.PublishCacheDeleteMessage(strconv.Itoa(sid))

		}()
		c.JSON(http.StatusOK, gin.H{"message": "ok", "id": id})
	}

}

func CreatePessimisticOrder(c *gin.Context) {
	sid, err := strconv.Atoi(c.Param("sid"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	id := service.CreateOrderWithPessimisticLock(sid)
	if id == -1 {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "fail to create", "id": sid})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "ok", "id": id})
	}
}

func CreateOptimisticOrder(c *gin.Context) {
	sid, err := strconv.Atoi(c.Param("sid"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}

	// rate limit
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if util.RateLimiter.Wait(ctx) != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "rate limiting"})
		return
	}

	remain := service.CreateOrderWithOptimisticLock(sid)
	if remain == -1 {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "fail to create", "remain": remain})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "ok", "remain": remain})
	}
}

// CreateOrderWithVerifiedUrl
// 带hash验证的订单创建
func CreateOrderWithVerifiedUrl(c *gin.Context) {
	sid, err1 := strconv.Atoi(c.Param("sid"))
	userId, err2 := strconv.Atoi(c.Param("userId"))
	verifyHash := c.Param("verifyHash")
	if err1 != nil || err2 != nil || verifyHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	remain, err := service.CreateOrderWithVerifiedUrl(sid, userId, verifyHash)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok", "remain": remain})
}

// CreateOrderWithVerifiedUrlAndLimit
// 带hash验证和单用户限制的订单创建
func CreateOrderWithVerifiedUrlAndLimit(c *gin.Context) {
	sid, err1 := strconv.Atoi(c.Param("sid"))
	userId, err2 := strconv.Atoi(c.Param("userId"))
	verifyHash := c.Param("verifyHash")
	if err1 != nil || err2 != nil || verifyHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	count := service.UserCountSelfIncrement(userId)
	if count == -1 {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "server error."})
		return
	}
	isBanned := service.GetUserIsBanned(userId)
	if isBanned {
		c.JSON(http.StatusOK, gin.H{"message": "超过限制下单数，请一个小时后再试"})
		return
	}

	remain, err := service.CreateOrderWithVerifiedUrl(sid, userId, verifyHash)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "remain": remain})
}

// GetVerifyHash
// 为抢购接口加盐
func GetVerifyHash(c *gin.Context) {
	sid, err1 := strconv.Atoi(c.Param("sid"))
	userId, err2 := strconv.Atoi(c.Param("userId"))
	if err1 != nil || err2 != nil {
		var messageBuilder strings.Builder
		if err1 != nil {
			messageBuilder.WriteString(err1.Error())
		}
		if err2 != nil {
			messageBuilder.WriteString(", " + err2.Error())
		}
		c.JSON(http.StatusBadRequest, gin.H{"message": messageBuilder.String()})
		return
	}

	hashCode, err := service.GetVerifyHashForSeckillURL(sid, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "success", "code": hashCode})
	}

}

func GetStockByDB(c *gin.Context) {
	sid, err := strconv.Atoi(c.Param("sid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	count := service.GetStockCountByDB(sid)
	if count == -1 {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库找不到该库存"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count, "message": "OK"})
}

func GetStockByCache(c *gin.Context) {
	sid, err := strconv.Atoi(c.Param("sid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	count := service.GetStockCountByCache(sid)
	if count == -1 {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "缓存未有该库存热点数据"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count, "message": "OK"})
}
