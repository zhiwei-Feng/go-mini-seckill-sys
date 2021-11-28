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
	dsn := "root:root@tcp(127.0.0.1:3306)/seckill?charset=utf8mb4&parseTime=True&loc=Local"
	dbcon, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	sqlDb, err := dbcon.DB()
	if err != nil {
		panic(err.Error())
	}
	sqlDb.SetMaxIdleConns(10)
	sqlDb.SetMaxOpenConns(100)
	sqlDb.SetConnMaxIdleTime(time.Minute)
	db.DbConn = dbcon

	r := gin.Default()
	r.GET("/createWrongOrder/:sid", controller.CreateWrongOrder)

	r.Run(":8888")
}
