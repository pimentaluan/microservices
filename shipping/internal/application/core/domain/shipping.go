package domain

import "time"

type ShippingItem struct {
	ProductCode string `json:"product_code"`
	Quantity    int32  `json:"quantity"`
}

type Shipping struct {
	ID                   int64          `json:"id"`
	OrderID              int64          `json:"order_id"`
	DeliveryForecastDays int32          `json:"delivery_forecast_days"`
	Items                []ShippingItem `json:"items"`
	CreatedAt            int64          `json:"created_at"`
}

func NewShipping(orderID int64, items []ShippingItem) Shipping {
	return Shipping{
		OrderID:   orderID,
		Items:     items,
		CreatedAt: time.Now().Unix(),
	}
}

func (s *Shipping) TotalItems() int32 {
	var totalItems int32

	for _, item := range s.Items {
		totalItems += item.Quantity
	}

	return totalItems
}
