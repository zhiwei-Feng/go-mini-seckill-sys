package service

import (
	"log"
	"mini-seckill/dao"
	"mini-seckill/db"
	"mini-seckill/util"
	"strconv"
)

//func UpdateStockById(stock domain.Stock) int {
//	return dao.UpdateStockByPk(stock)
//}
//
//func GetStockById(id int) domain.Stock {
//	return dao.SelectStockByPk(id)
//}

func GetStockCountByDB(id int) int {
	stock, err := dao.SelectStockByPk(db.DbConn, id)
	if err != nil {
		log.Println(err.Error())
		return -1
	}

	return stock.Count - stock.Sale
}

func GetStockCountByCache(id int) int {
	key := STOCK_COUNT + "_" + strconv.Itoa(id)
	valStr, err := util.GetRedisStringVal(key)
	val, convErr := strconv.Atoi(valStr)
	if err != nil || convErr != nil {
		log.Println(err.Error(), convErr.Error(), "cache未命中")
		stock, err := dao.SelectStockByPk(db.DbConn, id)
		if err != nil {
			log.Println(err.Error())
			return -1
		}
		err = util.SetRedisStringVal(key, strconv.Itoa(stock.Count-stock.Sale))
		if err != nil {
			log.Println(err.Error())
			return -1
		}
		return stock.Count - stock.Sale
		//return -1
	}
	return val
}

func DeleteStockCountCache(id int) bool {
	cacheKey := STOCK_COUNT + "_" + strconv.Itoa(id)
	err := util.DelRedisKey(cacheKey)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}
