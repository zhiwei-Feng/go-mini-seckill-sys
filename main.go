package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"mini-seckill/controller"
	"mini-seckill/db"
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

	// launch gin and config related handler
	r := gin.Default()
	r.GET("/createWrongOrder/:sid", controller.CreateWrongOrder)
	r.GET("/createOptimisticOrder/:sid", controller.CreateOptimisticOrder)
	r.GET("/createOrderWithVerifiedUrl/:sid/:userId/:verifyHash", controller.CreateOrderWithVerifiedUrl)
	r.GET("/getVerityHash/:sid/:userId", controller.GetVerifyHash)

	_ = r.Run(":8888")
}
