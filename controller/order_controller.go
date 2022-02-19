package controller

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"mini-seckill/domain"
	"mini-seckill/message"
	"mini-seckill/service"
	"mini-seckill/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CreateOrderWithMq 下单接口：异步订单创建
//
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

	hasOrder, err := service.CheckOrderInCache(sid, userId)
	if err != nil || hasOrder {
		log.Println("该用户已经抢购过")
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
				log.Println("等待1s后再删除")
				time.Sleep(time.Second)
				res := service.DeleteStockCountCache(sid)
				if !res {
					log.Println("再删除失败")
				} else {
					log.Println("再删除成功")
				}
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
			log.Println("等待1s后再删除")
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
			log.Println("再删除失败，放入消息队列重试")
			err := message.PublishCacheDeleteMessage(strconv.Itoa(sid))
			if err != nil {
				log.Println("发布消息失败")
			}
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
		log.Println("rate limited.")
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
		log.Println("请求参数错误")
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
		log.Println("请求参数错误")
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
		log.Println("获取验证hash失败，原因：", err.Error())
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
