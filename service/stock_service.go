package service

import "mini-seckill/domain"

type StockService interface {
	// GetStockById
	// @param id stock ID
	// @return domain.Stock
	GetStockById(id int) domain.Stock

	// UpdateStockById
	// @param stock old domain.Stock info(include ID)
	// @return stock ID
	UpdateStockById(stock domain.Stock) int
}
