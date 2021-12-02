package service

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"mini-seckill/dao"
	"mini-seckill/db"
	"mini-seckill/domain"
	"time"
)

// CreateWrongOrder
// @param sid stock ID
// @return order ID
func CreateWrongOrder(sid int) int {
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
