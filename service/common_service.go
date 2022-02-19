package service

import (
	"crypto/md5"
	"fmt"
	"log"
	"mini-seckill/dao"
	"mini-seckill/db"
	"mini-seckill/util"
	"strconv"
)

// SALT 加密盐
// 实际场景下，SALT会更长且会动态变化，无法直接硬编码在程序中
var SALT = "go-mini-seckill-sys"

func GetVerifyHashForSeckillURL(sid, userId int) (string, error) {
	// todo: 验证是否在抢购时间内

	// 检查用户
	_, err := dao.SelectUserByPk(db.DbConn, uint64(userId))
	if err != nil {
		log.Println("用户不存在")
		return "", err
	}

	// 检查商品
	_, err = dao.SelectStockByPk(db.DbConn, sid)
	if err != nil {
		log.Println("商品不存在")
		return "", err
	}

	// generate hashcode
	hashcode := md5.Sum([]byte(SALT + fmt.Sprintf("%v %v", sid, userId)))
	hashVal := fmt.Sprintf("%x", hashcode)

	// store into redis
	hashKey := HASH_KEY + "_" + strconv.Itoa(sid) + "_" + strconv.Itoa(userId)
	err = util.SetRedisStringVal(hashKey, hashVal)
	if err != nil {
		log.Println("redis写入失败")
		return "", err
	}
	return hashVal, nil
}

const (
	HASH_KEY       = "seckill_hash"
	LIMIT_KEY      = "seckill_limit"
	USER_HAS_ORDER = "has_order"
	STOCK_COUNT    = "stock_count"
	ALLOW_COUNT    = 10
)
