package dao

import (
	"log"
	"mini-seckill/db"
	"mini-seckill/domain"
)

func SelectStockByPk(id int) domain.Stock {
	ans := domain.Stock{}
	result := db.DbConn.First(&ans, id)
	if result.Error != nil {
		log.Printf("Err: query failed. message:%v", result.Error.Error())
	}

	return ans
}

func UpdateStockByPk(stock domain.Stock) int {
	result := db.DbConn.Updates(stock)
	if result.Error != nil {
		log.Printf("Err: update failed. message:%v", result.Error.Error())
	}

	return int(stock.ID)
}
