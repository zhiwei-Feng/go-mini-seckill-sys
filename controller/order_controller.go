package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"mini-seckill/service"
	"mini-seckill/util"
	"net/http"
	"strconv"
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
