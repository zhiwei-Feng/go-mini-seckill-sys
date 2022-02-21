package service

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	"mini-seckill/config"
	"mini-seckill/dao"
	"mini-seckill/db"
	"mini-seckill/domain"
	"mini-seckill/util"
	"strconv"
	"time"
)

//func UpdateStockById(stock domain.Stock) int {
//	return dao.UpdateStockByPk(stock)
//}
//
//func GetStockById(id int) domain.Stock {
//	return dao.SelectStockByPk(id)
//}

// GetStock 获取库存
// 先查询缓存，查不到再查DB
func GetStock(stockId int) (int, error) {
	key := config.GenerateStockKey(stockId)
	c := context.Background()
	obj, err := util.RedisCli.Get(c, key).Bytes()
	if err == redis.Nil {
		// 查询DB
		stock, err := dao.SelectStockByPk(db.DbConn, stockId)
		if err != nil {
			log.Error().Err(err).Msg("DB查询失败")
			return -1, err
		}
		// 写入缓存
		val, err := json.Marshal(stock)
		if err != nil {
			log.Error().Err(err).Msg("json序列化异常")
			return stock.Count - stock.Sale, nil
		}
		err = util.RedisCli.Set(c, key, val, time.Hour).Err()
		if err != nil {
			log.Error().Err(err).Msg("error in redis")
		}
		return stock.Count - stock.Sale, nil

	} else if err != nil {
		log.Error().Err(err).Msg("error in redis")
		return -1, err
	}
	stock := &domain.Stock{}
	err = json.Unmarshal(obj, stock)
	if err != nil {
		log.Warn().Err(err).Msg("异常的stock缓存")
		return -1, err
	}
	return stock.Count - stock.Sale, nil
}

func GetStockCountByDB(id int) int {
	stock, err := dao.SelectStockByPk(db.DbConn, id)
	if err != nil {
		return -1
	}

	return stock.Count - stock.Sale
}

func GetStockCountByCache(id int) int {
	key := STOCK_COUNT + "_" + strconv.Itoa(id)
	valStr, err := util.GetRedisStringVal(key)
	val, convErr := strconv.Atoi(valStr)
	if err != nil || convErr != nil {
		stock, err := dao.SelectStockByPk(db.DbConn, id)
		if err != nil {
			return -1
		}
		err = util.SetRedisStringVal(key, strconv.Itoa(stock.Count-stock.Sale))
		if err != nil {
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
		return false
	}
	return true
}
