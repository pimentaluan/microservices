package ports

import "github.com/pimentaluan/microservices/order/internal/application/core/domain"

type DBPort interface {
	Get(id string) (domain.Order, error)
	Save(*domain.Order) error
	UpdateStatus(id int64, status string) error
}
