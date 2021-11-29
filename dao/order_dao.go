package dao

import (
	"gorm.io/gorm"
	"log"
	"mini-seckill/domain"
)

func InsertOrderSelective(db *gorm.DB, order domain.StockOrder) (int, error) {
	result := db.Create(&order)
	if result.Error != nil {
		log.Printf("Err: insert order failed. message:%v", result.Error.Error())
		return -1, result.Error
	}

	return int(order.ID), nil
}
