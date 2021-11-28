package service

import (
	"errors"
	"log"
	"mini-seckill/dao"
	"mini-seckill/domain"
	"time"
)

func CreateWrongOrder(sid int) int {
	stock, err := checkStock(sid)
	if err != nil {
		log.Println(err)
		return -1
	}
	saleStock(stock)
	id := createOrder(stock)
	return id
}

func checkStock(sid int) (domain.Stock, error) {
	stock := dao.SelectStockByPk(sid)
	if stock.Name == "" {
		return stock, errors.New("stock don't exist")
	}
	if stock.Sale == stock.Count {
		return stock, errors.New("stockouts")
	}
	return stock, nil
}

func saleStock(stock domain.Stock) int {
	stock.Sale = stock.Sale - 1
	return UpdateStockById(stock)
}

func createOrder(stock domain.Stock) int {
	order := domain.StockOrder{}
	order.Sid = int(stock.ID)
	order.Name = stock.Name
	order.CreateTime = time.Now()
	// invoke dao to create
	id := dao.InsertOrderSelective(order)
	return id
}
