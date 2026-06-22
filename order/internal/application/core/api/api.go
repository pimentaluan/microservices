package api

import (
	"github.com/pimentaluan/microservices/order/internal/application/core/domain"
	"github.com/pimentaluan/microservices/order/internal/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Application struct {
	db      ports.DBPort
	payment ports.PaymentPort
}

func NewApplication(db ports.DBPort, payment ports.PaymentPort) *Application {
	return &Application{
		db:      db,
		payment: payment,
	}
}

func (a Application) PlaceOrder(order domain.Order) (domain.Order, error) {
	err := a.db.Save(&order)

	if err != nil {
		return domain.Order{}, err
	}

	if order.TotalItems() > 50 {
		if updateErr := a.db.UpdateStatus(order.ID, "Canceled"); updateErr != nil {
			return domain.Order{}, updateErr
		}

		return domain.Order{}, status.Error(codes.InvalidArgument, "orders over 50 items are not allowed")
	}

	paymentErr := a.payment.Charge(&order)

	if paymentErr != nil {
		if updateErr := a.db.UpdateStatus(order.ID, "Canceled"); updateErr != nil {
			return domain.Order{}, updateErr
		}

		return domain.Order{}, paymentErr
	}

	if updateErr := a.db.UpdateStatus(order.ID, "Paid"); updateErr != nil {
		return domain.Order{}, updateErr
	}

	order.Status = "Paid"

	return order, nil
}
