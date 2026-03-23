package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	inventoryV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/inventory/v1"
)

const grpcPort = 50051

type inventoryService struct {
	inventoryV1.UnimplementedInventoryServiceServer

	mu    sync.RWMutex
	parts map[string]*inventoryV1.Part
}

func newInventoryService() *inventoryService {
	s := &inventoryService{
		parts: make(map[string]*inventoryV1.Part),
	}
	s.seedData()
	return s
}

func (s *inventoryService) seedData() {
	now := timestamppb.New(time.Now())

	parts := []*inventoryV1.Part{
		{
			Uuid:          uuid.NewString(),
			Name:          "Raptor Engine V2",
			Description:   "Основной двигатель для межпланетных перелётов",
			Price:         1500000.00,
			StockQuantity: 10,
			Category:      inventoryV1.Category_CATEGORY_ENGINE,
			Dimensions: &inventoryV1.Dimensions{
				Length: 380,
				Width:  130,
				Height: 130,
				Weight: 1600,
			},
			Manufacturer: &inventoryV1.Manufacturer{
				Name:    "SpaceX",
				Country: "USA",
				Website: "https://spacex.com",
			},
			Tags: []string{"engine", "raptor", "methane"},
			Metadata: map[string]*inventoryV1.Value{
				"thrust_kn": {Kind: &inventoryV1.Value_DoubleValue{DoubleValue: 2300}},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Uuid:          uuid.NewString(),
			Name:          "Liquid Methane Tank",
			Description:   "Топливный бак для жидкого метана",
			Price:         250000.00,
			StockQuantity: 25,
			Category:      inventoryV1.Category_CATEGORY_FUEL,
			Dimensions: &inventoryV1.Dimensions{
				Length: 1200,
				Width:  400,
				Height: 400,
				Weight: 5000,
			},
			Manufacturer: &inventoryV1.Manufacturer{
				Name:    "RocketLab",
				Country: "New Zealand",
				Website: "https://rocketlabusa.com",
			},
			Tags: []string{"fuel", "methane", "tank"},
			Metadata: map[string]*inventoryV1.Value{
				"capacity_liters": {Kind: &inventoryV1.Value_Int64Value{Int64Value: 120000}},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Uuid:          uuid.NewString(),
			Name:          "Reinforced Porthole",
			Description:   "Усиленный иллюминатор для наблюдения из космоса",
			Price:         75000.00,
			StockQuantity: 50,
			Category:      inventoryV1.Category_CATEGORY_PORTHOLE,
			Dimensions: &inventoryV1.Dimensions{
				Length: 60,
				Width:  60,
				Height: 15,
				Weight: 45,
			},
			Manufacturer: &inventoryV1.Manufacturer{
				Name:    "Airbus Defence",
				Country: "Germany",
				Website: "https://airbus.com",
			},
			Tags: []string{"porthole", "glass", "observation"},
			Metadata: map[string]*inventoryV1.Value{
				"uv_protection": {Kind: &inventoryV1.Value_BoolValue{BoolValue: true}},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Uuid:          uuid.NewString(),
			Name:          "Delta Wing Module",
			Description:   "Модуль крыла для атмосферного входа",
			Price:         320000.00,
			StockQuantity: 15,
			Category:      inventoryV1.Category_CATEGORY_WING,
			Dimensions: &inventoryV1.Dimensions{
				Length: 800,
				Width:  300,
				Height: 50,
				Weight: 1200,
			},
			Manufacturer: &inventoryV1.Manufacturer{
				Name:    "Boeing",
				Country: "USA",
				Website: "https://boeing.com",
			},
			Tags: []string{"wing", "delta", "atmospheric"},
			Metadata: map[string]*inventoryV1.Value{
				"max_temperature_c": {Kind: &inventoryV1.Value_DoubleValue{DoubleValue: 1650}},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, p := range parts {
		s.parts[p.Uuid] = p
	}
}

func (s *inventoryService) GetPart(_ context.Context, req *inventoryV1.GetPartRequest) (*inventoryV1.GetPartResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	part, ok := s.parts[req.GetUuid()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "part with UUID %s not found", req.GetUuid())
	}

	return &inventoryV1.GetPartResponse{Part: part}, nil
}

func (s *inventoryService) ListParts(_ context.Context, req *inventoryV1.ListPartsRequest) (*inventoryV1.ListPartsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Собираем все детали
	result := make([]*inventoryV1.Part, 0, len(s.parts))
	for _, p := range s.parts {
		result = append(result, p)
	}

	filter := req.GetFilter()
	if filter == nil {
		return &inventoryV1.ListPartsResponse{Parts: result}, nil
	}

	// Фильтрация по UUID
	if len(filter.Uuids) > 0 {
		uuidSet := make(map[string]struct{}, len(filter.Uuids))
		for _, u := range filter.Uuids {
			uuidSet[u] = struct{}{}
		}
		result = filterParts(result, func(p *inventoryV1.Part) bool {
			_, ok := uuidSet[p.Uuid]
			return ok
		})
	}

	// Фильтрация по имени
	if len(filter.Names) > 0 {
		nameSet := make(map[string]struct{}, len(filter.Names))
		for _, n := range filter.Names {
			nameSet[n] = struct{}{}
		}
		result = filterParts(result, func(p *inventoryV1.Part) bool {
			_, ok := nameSet[p.Name]
			return ok
		})
	}

	// Фильтрация по категории
	if len(filter.Categories) > 0 {
		catSet := make(map[inventoryV1.Category]struct{}, len(filter.Categories))
		for _, c := range filter.Categories {
			catSet[c] = struct{}{}
		}
		result = filterParts(result, func(p *inventoryV1.Part) bool {
			_, ok := catSet[p.Category]
			return ok
		})
	}

	// Фильтрация по стране производителя
	if len(filter.ManufacturerCountries) > 0 {
		countrySet := make(map[string]struct{}, len(filter.ManufacturerCountries))
		for _, c := range filter.ManufacturerCountries {
			countrySet[c] = struct{}{}
		}
		result = filterParts(result, func(p *inventoryV1.Part) bool {
			if p.Manufacturer == nil {
				return false
			}
			_, ok := countrySet[p.Manufacturer.Country]
			return ok
		})
	}

	// Фильтрация по тегам
	if len(filter.Tags) > 0 {
		tagSet := make(map[string]struct{}, len(filter.Tags))
		for _, t := range filter.Tags {
			tagSet[t] = struct{}{}
		}
		result = filterParts(result, func(p *inventoryV1.Part) bool {
			for _, tag := range p.Tags {
				if _, ok := tagSet[tag]; ok {
					return true
				}
			}
			return false
		})
	}

	return &inventoryV1.ListPartsResponse{Parts: result}, nil
}

func filterParts(parts []*inventoryV1.Part, predicate func(*inventoryV1.Part) bool) []*inventoryV1.Part {
	filtered := make([]*inventoryV1.Part, 0)
	for _, p := range parts {
		if predicate(p) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	service := newInventoryService()
	inventoryV1.RegisterInventoryServiceServer(s, service)

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
