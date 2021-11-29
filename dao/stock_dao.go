package dao

import (
	"gorm.io/gorm"
	"log"
	"mini-seckill/domain"
)

func SelectStockByPk(db *gorm.DB, id int) (domain.Stock, error) {
	ans := domain.Stock{}
	result := db.First(&ans, id)
	if result.Error != nil {
		log.Printf("Err: query failed. message:%v", result.Error.Error())
		return ans, result.Error
	}

	return ans, nil
}

func UpdateStockByPk(db *gorm.DB, stock domain.Stock) int {
	result := db.Updates(stock)
	if result.Error != nil {
		log.Printf("Err: update failed. message:%v", result.Error.Error())
	}

	return int(stock.ID)
}
