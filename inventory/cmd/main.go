package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	inventoryApi "github.com/AlexKostromin/microsrv/inventory/internal/api/inventory/v1"
	inventoryRepo "github.com/AlexKostromin/microsrv/inventory/internal/repository/inventory"
	inventoryService "github.com/AlexKostromin/microsrv/inventory/internal/service/inventory"
	inventoryV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/inventory/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const grpcPort = 50051

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	repo := inventoryRepo.New()
	svc := inventoryService.NewService(repo)
	api := inventoryApi.NewApi(svc)

	inventoryV1.RegisterInventoryServiceServer(s, api)

	reflection.Register(s)

	go func() {
		log.Printf("gRPC InventoryService listening on :%d\n", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down InventoryService...")
	s.GracefulStop()
	log.Println("InventoryService stopped")
}
