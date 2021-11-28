package service

import (
	"mini-seckill/dao"
	"mini-seckill/domain"
)

func UpdateStockById(stock domain.Stock) int {
	return dao.UpdateStockByPk(stock)
}

func GetStockById(id int) domain.Stock {
	return dao.SelectStockByPk(id)
}
