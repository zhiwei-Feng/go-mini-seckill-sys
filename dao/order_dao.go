package dao

import (
	"gorm.io/gorm"
	"log"
	"mini-seckill/domain"
)

func InsertOrder(db *gorm.DB, order domain.StockOrder) (int, error) {
	result := db.Create(&order)
	if result.Error != nil {
		log.Printf("Err: insert order failed. message:%v", result.Error.Error())
		return -1, result.Error
	}

	return int(order.ID), nil
}

func CountOrderByIdAndUserId(db *gorm.DB, stockId, userId int) (int, error) {
	var count int64
	res := db.Where(&domain.StockOrder{Sid: stockId, UserId: userId}).Count(&count)
	if res != nil {
		return -1, nil
	}
	return int(count), nil
}
