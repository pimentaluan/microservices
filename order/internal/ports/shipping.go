package ports

import "github.com/pimentaluan/microservices/order/internal/application/core/domain"

type ShippingPort interface {
	Create(order *domain.Order) (int32, error)
}
