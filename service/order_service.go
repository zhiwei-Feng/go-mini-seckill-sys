package service

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"mini-seckill/dao"
	"mini-seckill/db"
	"mini-seckill/domain"
	"mini-seckill/util"
	"strconv"
	"time"
)

func CreateOrderWithMq(sid int, userId int) int {
	var remaining int
	err := db.DbConn.Transaction(func(tx *gorm.DB) error {
		stock, err := dao.SelectStockByPk(tx, sid)
		if err != nil {
			return err
		}
		if stock.Sale == stock.Count {
			log.Println("StockOuts")
			return errors.New("StockOuts")
		}

		_, err = dao.UpdateStockByPkWithOptimistic(tx, stock)
		if err != nil {
			log.Println("乐观锁并发控制")
			return err
		}
		remaining = stock.Count - stock.Sale

		order := domain.StockOrder{}
		order.Sid = int(stock.ID)
		order.Name = stock.Name
		order.CreateTime = time.Now()
		order.UserId = userId
		_, err = dao.InsertOrderSelective(tx, order)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return -1
	}

	DeleteStockCountCache(sid)
	key := USER_HAS_ORDER + "_" + strconv.Itoa(sid)
	err = util.SetAdd(key, strconv.Itoa(userId))
	if err != nil {
		return -1
	}
	return remaining
}

// CreateOrderWithPessimisticLock
// @param sid stock ID
// @return order ID
func CreateOrderWithPessimisticLock(sid int) int {
	id := -1
	err := db.DbConn.Transaction(func(tx *gorm.DB) error {
		// add lock
		stock, err := dao.SelectStockByPk(tx.Clauses(clause.Locking{Strength: "UPDATE"}), sid)
		if err != nil {
			return err
		}
		if stock.Sale == stock.Count {
			log.Println("StockOuts")
			return errors.New("StockOuts")
		}

		_, err = dao.UpdateStockByPk(tx, stock)
		if err != nil {
			log.Println("悲观锁条件下更新失败.")
			return err
		}

		// create order
		order := domain.StockOrder{}
		order.Sid = int(stock.ID)
		order.Name = stock.Name
		order.CreateTime = time.Now()
		id, err = dao.InsertOrderSelective(tx, order)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return -1
	}
	return id
}

// CreateOrderWithOptimisticLock
// @param sid stock id
// @return remaining inventory (stock.Count-stock.sale)
func CreateOrderWithOptimisticLock(sid int) int {
	var remaining int
	err := db.DbConn.Transaction(func(tx *gorm.DB) error {
		stock, err := dao.SelectStockByPk(tx, sid)
		if err != nil {
			return err
		}
		if stock.Sale == stock.Count {
			log.Println("StockOuts")
			return errors.New("StockOuts")
		}

		_, err = dao.UpdateStockByPkWithOptimistic(tx, stock)
		if err != nil {
			log.Println("乐观锁并发控制")
			return err
		}
		remaining = stock.Count - stock.Sale

		order := domain.StockOrder{}
		order.Sid = int(stock.ID)
		order.Name = stock.Name
		order.CreateTime = time.Now()
		_, err = dao.InsertOrderSelective(tx, order)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return -1
	}
	return remaining
}

func CreateOrderWithVerifiedUrl(sid, userId int, hashcode string) (int, error) {
	// todo: 验证是否在抢购时间内

	// 验证hashcode的合法性
	hashKey := HASH_KEY + "_" + strconv.Itoa(sid) + "_" + strconv.Itoa(userId)
	verifiedHash, err := util.GetRedisStringVal(hashKey)
	if err != nil {
		log.Println("hash获取失败, ", err.Error())
		return -1, err
	}

	if verifiedHash != hashcode {
		log.Println("hash验证失败")
		return -1, errors.New("错误的hashcode")
	}

	// 验证用户
	_, err = dao.SelectUserByPk(db.DbConn, uint64(userId))
	if err != nil {
		log.Println("用户不存在")
		return -1, err
	}

	// 验证商品
	stock, err := dao.SelectStockByPk(db.DbConn, sid)
	if err != nil {
		log.Println("商品不存在")
		return -1, err
	}

	// 乐观锁更新库存
	var remain int
	err = db.DbConn.Transaction(func(tx *gorm.DB) error {
		if stock.Sale == stock.Count {
			log.Println("StockOuts")
			return errors.New("StockOuts")
		}

		_, err = dao.UpdateStockByPkWithOptimistic(tx, stock)
		if err != nil {
			log.Println("乐观锁并发控制")
			return err
		}
		log.Printf("最大量:%v，卖出量:%v\n", stock.Count, stock.Sale+1)
		remain = stock.Count - stock.Sale - 1

		// create order
		order := domain.StockOrder{}
		order.Sid = int(stock.ID)
		order.Name = stock.Name
		order.CreateTime = time.Now()
		order.UserId = userId
		_, err = dao.InsertOrderSelective(tx, order)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return -1, err
	}

	return remain, nil
}

func CheckOrderInCache(sid, userId int) (bool, error) {
	key := USER_HAS_ORDER + "_" + strconv.Itoa(sid)
	log.Printf("检查用户Id：[%v] 是否抢购过商品Id：[%v] 检查Key：[%s]", userId, sid, key)
	res, err := util.IsMember(key, strconv.Itoa(userId))
	if err != nil {
		return false, err
	}
	return res, nil
}
