package ports

import "github.com/pimentaluan/microservices/shipping/internal/application/core/domain"

type DBPort interface {
	Save(*domain.Shipping) error
}
