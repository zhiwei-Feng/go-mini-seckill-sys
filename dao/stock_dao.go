package dao

import (
	"errors"
	"gorm.io/gorm"
	"log"
	"mini-seckill/domain"
)

func SelectStockByPk(db *gorm.DB, id int) (domain.Stock, error) {
	ans := domain.Stock{}
	result := db.First(&ans, id)
	if result.Error != nil {
		log.Printf("Err: query failed. message:%v", result.Error.Error())
		return ans, result.Error
	}

	return ans, nil
}

func UpdateStockByPk(db *gorm.DB, stock domain.Stock) (int, error) {
	result := db.Debug().
		Model(&stock).
		Updates(map[string]interface{}{
			"sale": gorm.Expr("sale + ?", 1),
		})
	if result.Error != nil || result.RowsAffected == 0 {
		err := result.Error
		if err == nil {
			err = errors.New("更新失败")
		}
		return -1, err
	}

	return int(stock.ID), nil
}

func UpdateStockByPkWithOptimistic(db *gorm.DB, stock domain.Stock) (int, error) {
	result := db.Debug().
		Model(&stock).
		Where("version = ?", stock.Version).Updates(map[string]interface{}{
		"sale":    gorm.Expr("sale + ?", 1),
		"version": gorm.Expr("version + ?", 1),
	})
	if result.Error != nil || result.RowsAffected == 0 {
		err := result.Error
		if err == nil {
			err = errors.New("乐观锁并发控制")
		}
		return -1, err
	}

	return int(stock.ID), nil
}
