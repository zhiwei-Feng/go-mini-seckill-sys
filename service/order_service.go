package service

type OrderService interface {
	// CreateWrongOrder
	// @param sid stock ID
	// @return order ID
	CreateWrongOrder(sid int) int
}
