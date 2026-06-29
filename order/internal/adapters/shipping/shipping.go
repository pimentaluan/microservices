package shipping

import (
	"context"
	"log"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	shippingpb "github.com/pimentaluan/microservices-proto/golang/shipping"
	"github.com/pimentaluan/microservices/order/internal/application/core/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Adapter struct {
	shipping shippingpb.ShippingClient
}

func NewAdapter(shippingServiceURL string) (*Adapter, error) {
	var opts []grpc.DialOption

	opts = append(opts,
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
			grpc_retry.WithCodes(codes.Unavailable, codes.ResourceExhausted),
			grpc_retry.WithMax(5),
			grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Second)),
		)),
	)
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient(
		shippingServiceURL,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	client := shippingpb.NewShippingClient(conn)

	return &Adapter{shipping: client}, nil
}

func (a *Adapter) Create(order *domain.Order) (int32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var items []*shippingpb.ShippingItem

	for _, item := range order.OrderItems {
		items = append(items, &shippingpb.ShippingItem{
			ProductCode: item.ProductCode,
			Quantity:    item.Quantity,
		})
	}

	response, err := a.shipping.Create(
		ctx,
		&shippingpb.CreateShippingRequest{
			OrderId: order.ID,
			Items:   items,
		},
	)
	if status.Code(err) == codes.DeadlineExceeded {
		log.Printf("timeout ao chamar o microsservico shipping para o pedido %d", order.ID)
	}
	if err != nil {
		return 0, err
	}

	return response.DeliveryForecastDays, nil
}
