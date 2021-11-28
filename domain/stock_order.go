package domain

import (
	"time"
)

type StockOrder struct {
	ID         uint
	Sid        int
	Name       string
	CreateTime time.Time
}

func (receiver StockOrder) TableName() string {
	return "stock_order"
}
