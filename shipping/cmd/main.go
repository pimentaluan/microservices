package main

import (
	"log"

	"github.com/pimentaluan/microservices/shipping/config"
	"github.com/pimentaluan/microservices/shipping/internal/adapters/db"
	grpcAdapter "github.com/pimentaluan/microservices/shipping/internal/adapters/grpc"
	"github.com/pimentaluan/microservices/shipping/internal/application/core/api"
)

func main() {
	dbAdapter, err := db.NewAdapter(config.GetDataSourceURL())
	if err != nil {
		log.Fatalf("erro ao iniciar banco: %v", err)
	}

	application := api.NewApplication(dbAdapter)
	adapter := grpcAdapter.NewAdapter(application, config.GetApplicationPort())
	adapter.Run()
}
