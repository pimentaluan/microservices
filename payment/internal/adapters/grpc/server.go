package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	paymentpb "github.com/pimentaluan/microservices-proto/golang/payment"
	"github.com/pimentaluan/microservices/payment/config"
	"github.com/pimentaluan/microservices/payment/internal/application/core/domain"
	"github.com/pimentaluan/microservices/payment/internal/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Adapter struct {
	api  ports.APIPort
	port int
	paymentpb.UnimplementedPaymentServer
}

func NewAdapter(api ports.APIPort, port int) *Adapter {
	return &Adapter{api: api, port: port}
}

func (a Adapter) Create(ctx context.Context, request *paymentpb.CreatePaymentRequest) (*paymentpb.CreatePaymentResponse, error) {
	newPayment := domain.NewPayment(request.UserId, request.OrderId, request.TotalPrice)

	result, err := a.api.CreatePayment(newPayment)
	if err != nil {
		return nil, err
	}

	return &paymentpb.CreatePaymentResponse{
		PaymentId: result.ID,
		BillId:    result.BillID,
	}, nil
}

func (a Adapter) Run() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Fatalf("erro ao escutar porta %d: %v", a.port, err)
	}

	grpcServer := grpc.NewServer()
	paymentpb.RegisterPaymentServer(grpcServer, a)

	if config.GetEnv() == "development" {
		reflection.Register(grpcServer)
	}

	log.Printf("servidor gRPC rodando na porta %d", a.port)

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("erro ao iniciar servidor gRPC: %v", err)
	}
}
