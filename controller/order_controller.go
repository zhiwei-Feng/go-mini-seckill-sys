package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"mini-seckill/service"
	"mini-seckill/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func CreateWrongOrder(c *gin.Context) {
	sid, err := strconv.Atoi(c.Param("sid"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	id := service.CreateWrongOrder(sid)
	if id == -1 {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "fail to create", "id": id})
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
