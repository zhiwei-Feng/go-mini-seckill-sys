package main

import (
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"mini-seckill/controller"
	"mini-seckill/db"
	"mini-seckill/message"
	"time"
)

func main() {
	// open database connection
	dsn := "root:root@tcp(127.0.0.1:3306)/seckill?charset=utf8mb4&parseTime=True&loc=Local"
	dbcon, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	sqlDb, err := dbcon.DB()
	if err != nil {
		panic(err.Error())
	}
	sqlDb.SetMaxIdleConns(100)
	sqlDb.SetMaxOpenConns(1000)
	sqlDb.SetConnMaxIdleTime(time.Minute * 30)
	db.DbConn = dbcon

	// start rabbitmq consumer
	//go message.ConsumerForCacheDeleteMessage()
	go message.ConsumerForOrderCreate()
	go message.ConsumerForStockCacheDelete()

	// launch gin and config related handler
	r := gin.Default()
	r.GET("/createPessimisticOrder/:sid", controller.CreatePessimisticOrder)
	r.GET("/createOptimisticOrder/:sid", controller.CreateOptimisticOrder)
	r.GET("/createOrderWithVerifiedUrl/:sid/:userId/:verifyHash", controller.CreateOrderWithVerifiedUrl)
	r.GET("/createOrderWithVerifiedUrlAndLimit/:sid/:userId/:verifyHash", controller.CreateOrderWithVerifiedUrlAndLimit)
	r.GET("/getVerityHash/:sid/:userId", controller.GetVerifyHash)
	r.GET("/getStockByDB/:sid", controller.GetStockByDB)
	r.GET("/getStockByCache/:sid", controller.GetStockByCache)
	r.GET("/createOrderWithCacheV1/:sid", controller.CreateOrderWithCacheV1)
	r.GET("/createOrderWithCacheV2/:sid", controller.CreateOrderWithCacheV2)
	r.GET("/createOrderWithCacheV3/:sid", controller.CreateOrderWithCacheV3)
	r.GET("/createOrderWithCacheV4/:sid", controller.CreateOrderWithCacheV4)
	r.GET("/createOrderWithMq/:sid/:userId", controller.CreateOrderWithMq)

	goodsseckill := r.Group("/goodsseckill")
	goodsseckill.GET("/v1", controller.GoodsSeckillV1)
	goodsseckill.GET("/v2", controller.GoodsSeckillV2)

	_ = endless.ListenAndServe(":8888", r)
}
