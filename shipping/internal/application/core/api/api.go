package api

import (
	"context"

	"github.com/pimentaluan/microservices/shipping/internal/application/core/domain"
	"github.com/pimentaluan/microservices/shipping/internal/ports"
)

type Application struct {
	db ports.DBPort
}

func NewApplication(db ports.DBPort) *Application {
	return &Application{db: db}
}

func (a Application) CreateShipping(_ context.Context, shipping domain.Shipping) (domain.Shipping, error) {
	totalItems := shipping.TotalItems()
	deliveryForecastDays := int32(1)

	if totalItems > 0 {
		deliveryForecastDays += (totalItems - 1) / 5
	}

	shipping.DeliveryForecastDays = deliveryForecastDays

	err := a.db.Save(&shipping)
	if err != nil {
		return domain.Shipping{}, err
	}

	return shipping, nil
}
