package payment

import (
	"context"

	paymentpb "github.com/pimentaluan/microservices-proto/golang/payment"
	"github.com/pimentaluan/microservices/order/internal/application/core/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Adapter struct {
	payment paymentpb.PaymentClient
}

func NewAdapter(paymentServiceURL string) (*Adapter, error) {
	conn, err := grpc.NewClient(
		paymentServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := paymentpb.NewPaymentClient(conn)

	return &Adapter{payment: client}, nil
}

func (a *Adapter) Charge(order *domain.Order) error {
	_, err := a.payment.Create(
		context.Background(),
		&paymentpb.CreatePaymentRequest{
			UserId:     order.CustomerID,
			OrderId:    order.ID,
			TotalPrice: order.TotalPrice(),
		},
	)

	return err
}
