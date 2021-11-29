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

		stock.Sale += 1
		dao.UpdateStockByPk(tx, stock)
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
