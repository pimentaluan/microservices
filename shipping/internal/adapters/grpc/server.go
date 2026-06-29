package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	shippingpb "github.com/pimentaluan/microservices-proto/golang/shipping"
	"github.com/pimentaluan/microservices/shipping/config"
	"github.com/pimentaluan/microservices/shipping/internal/application/core/domain"
	"github.com/pimentaluan/microservices/shipping/internal/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type Adapter struct {
	api  ports.APIPort
	port int
	shippingpb.UnimplementedShippingServer
}

func NewAdapter(api ports.APIPort, port int) *Adapter {
	return &Adapter{api: api, port: port}
}

func (a Adapter) Create(ctx context.Context, request *shippingpb.CreateShippingRequest) (*shippingpb.CreateShippingResponse, error) {
	var items []domain.ShippingItem

	for _, item := range request.Items {
		items = append(items, domain.ShippingItem{
			ProductCode: item.ProductCode,
			Quantity:    item.Quantity,
		})
	}

	newShipping := domain.NewShipping(request.OrderId, items)

	result, err := a.api.CreateShipping(ctx, newShipping)
	if err != nil {
		if code := status.Code(err); code != codes.Unknown {
			return nil, err
		}

		return nil, status.Errorf(codes.Internal, "failed to create shipping: %v", err)
	}

	return &shippingpb.CreateShippingResponse{
		ShippingId:           result.ID,
		DeliveryForecastDays: result.DeliveryForecastDays,
	}, nil
}

func (a Adapter) Run() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Fatalf("erro ao escutar porta %d: %v", a.port, err)
	}

	grpcServer := grpc.NewServer()
	shippingpb.RegisterShippingServer(grpcServer, a)

	if config.GetEnv() == "development" {
		reflection.Register(grpcServer)
	}

	log.Printf("servidor gRPC rodando na porta %d", a.port)

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("erro ao iniciar servidor gRPC: %v", err)
	}
}
