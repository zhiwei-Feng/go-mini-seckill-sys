package dao

import (
	"log"
	"mini-seckill/db"
	"mini-seckill/domain"
)

func InsertOrderSelective(order domain.StockOrder) (id int) {
	result := db.DbConn.Create(&order)
	if result.Error != nil {
		log.Printf("Err: insert order failed. message:%v", result.Error.Error())
		return -1
	}

	return int(order.ID)
}
