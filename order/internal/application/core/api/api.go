package api

import (
	"github.com/pimentaluan/microservices/order/internal/application/core/domain"
	"github.com/pimentaluan/microservices/order/internal/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Application struct {
	db       ports.DBPort
	payment  ports.PaymentPort
	shipping ports.ShippingPort
}

func NewApplication(db ports.DBPort, payment ports.PaymentPort, shipping ports.ShippingPort) *Application {
	return &Application{
		db:       db,
		payment:  payment,
		shipping: shipping,
	}
}

func (a Application) PlaceOrder(order domain.Order) (domain.Order, error) {
	for _, item := range order.OrderItems {
		exists, err := a.db.ProductExists(item.ProductCode)
		if err != nil {
			return domain.Order{}, err
		}

		if !exists {
			return domain.Order{}, status.Errorf(codes.NotFound, "product %s was not found in inventory", item.ProductCode)
		}
	}

	if order.TotalItems() > 50 {
		return domain.Order{}, status.Error(codes.InvalidArgument, "orders over 50 items are not allowed")
	}

	err := a.db.Save(&order)

	if err != nil {
		return domain.Order{}, err
	}

	paymentErr := a.payment.Charge(&order)

	if paymentErr != nil {
		if updateErr := a.db.UpdateStatus(order.ID, "Canceled"); updateErr != nil {
			return domain.Order{}, updateErr
		}

		return domain.Order{}, paymentErr
	}

	_, shippingErr := a.shipping.Create(&order)
	if shippingErr != nil {
		if updateErr := a.db.UpdateStatus(order.ID, "Canceled"); updateErr != nil {
			return domain.Order{}, updateErr
		}

		return domain.Order{}, shippingErr
	}

	if updateErr := a.db.UpdateStatus(order.ID, "Paid"); updateErr != nil {
		return domain.Order{}, updateErr
	}

	order.Status = "Paid"

	return order, nil
}
