package view

type SeckillReq struct {
	StockId    int    `form:"stockId"`
	UserId     int    `form:"userId"`
	VerifyHash string `form:"hash"`
}
