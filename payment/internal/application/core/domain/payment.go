package domain

import "time"

type Payment struct {
	ID         int64   `json:"id"`
	UserID     int64   `json:"user_id"`
	OrderID    int64   `json:"order_id"`
	BillID     int64   `json:"bill_id"`
	TotalPrice float32 `json:"total_price"`
	CreatedAt  int64   `json:"created_at"`
}

func NewPayment(userID, orderID int64, totalPrice float32) Payment {
	return Payment{
		UserID:     userID,
		OrderID:    orderID,
		TotalPrice: totalPrice,
		CreatedAt:  time.Now().Unix(),
	}
}
