package ports

import "github.com/pimentaluan/microservices/payment/internal/application/core/domain"

type DBPort interface {
	Save(*domain.Payment) error
}
