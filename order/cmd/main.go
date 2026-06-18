package main

import (
	"log"

	"github.com/pimentaluan/microservices/order/config"
	"github.com/pimentaluan/microservices/order/internal/adapters/db"
	grpcAdapter "github.com/pimentaluan/microservices/order/internal/adapters/grpc"
	paymentAdapter "github.com/pimentaluan/microservices/order/internal/adapters/payment"
	"github.com/pimentaluan/microservices/order/internal/application/core/api"
)

func main() {
	dbAdapter, err := db.NewAdapter(config.GetDataSourceURL())

	if err != nil {
		log.Fatalf("erro ao iniciar banco: %v", err)
	}

	paymentStub, err := paymentAdapter.NewAdapter(config.GetPaymentServiceURL())

	if err != nil {
		log.Fatalf("erro ao iniciar cliente de pagamento: %v", err)
	}

	application := api.NewApplication(dbAdapter, paymentStub)

	adapter := grpcAdapter.NewAdapter(
		application,
		config.GetApplicationPort(),
	)

	adapter.Run()
}
