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
	res := db.Model(&domain.StockOrder{}).Where("sid = ? AND user_id = ?", stockId, userId).Count(&count)
	if res.Error != nil {
		return -1, res.Error
	}
	return int(count), nil
}
