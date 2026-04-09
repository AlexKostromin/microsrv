package inventory

import (
	"sync"
	"time"

	"github.com/AlexKostromin/microsrv/inventory/internal/repository"
	"github.com/AlexKostromin/microsrv/inventory/internal/repository/model"

	"github.com/google/uuid"
)

type Repository struct {
	mu    sync.RWMutex
	parts map[string]*model.Part
}

var _ repository.InventoryRepository = (*Repository)(nil)

func New() *Repository {
	now := time.Now()
	r := &Repository{
		parts: make(map[string]*model.Part),
	}
	parts := []*model.Part{
		{
			UUID:          uuid.NewString(),
			Name:          "Raptor Engine V2",
			Description:   "Основной двигатель для межпланетных перелётов",
			Price:         1500000.00,
			StockQuantity: 10,
			Category:      model.CategoryEngine,
			Dimensions: model.Dimensions{
				Length: 380,
				Width:  130,
				Height: 130,
				Weight: 1600,
			},
			Manufacturer: model.Manufacturer{
				Name:    "SpaceX",
				Country: "USA",
				Website: "https://spacex.com",
			},
			Tags: []string{"engine", "raptor", "methane"},
			Metadata: map[string]interface{}{
				"thrust_kn": 2300,
			},
			CreatedAt: now,
			UpdatedAt: &now,
		},
		{
			UUID:          uuid.NewString(),
			Name:          "Liquid Methane Tank",
			Description:   "Топливный бак для жидкого метана",
			Price:         250000.00,
			StockQuantity: 25,
			Category:      model.CategoryFuel,
			Dimensions: model.Dimensions{
				Length: 1200,
				Width:  400,
				Height: 400,
				Weight: 5000,
			},
			Manufacturer: model.Manufacturer{
				Name:    "RocketLab",
				Country: "New Zealand",
				Website: "https://rocketlabusa.com",
			},
			Tags: []string{"fuel", "methane", "tank"},
			Metadata: map[string]interface{}{
				"capacity_liters": 120000,
			},
			CreatedAt: now,
			UpdatedAt: &now,
		},
		{
			UUID:          uuid.NewString(),
			Name:          "Reinforced Porthole",
			Description:   "Усиленный иллюминатор для наблюдения из космоса",
			Price:         75000.00,
			StockQuantity: 50,
			Category:      model.CategoryPorthole,
			Dimensions: model.Dimensions{
				Length: 60,
				Width:  60,
				Height: 15,
				Weight: 45,
			},
			Manufacturer: model.Manufacturer{
				Name:    "Airbus Defence",
				Country: "Germany",
				Website: "https://airbus.com",
			},
			Tags: []string{"porthole", "glass", "observation"},
			Metadata: map[string]interface{}{
				"uv_protection": true,
			},
			CreatedAt: now,
			UpdatedAt: &now,
		},
		{
			UUID:          uuid.NewString(),
			Name:          "Delta Wing Module",
			Description:   "Модуль крыла для атмосферного входа",
			Price:         320000.00,
			StockQuantity: 15,
			Category:      model.CategoryWing,
			Dimensions: model.Dimensions{
				Length: 800,
				Width:  300,
				Height: 50,
				Weight: 1200,
			},
			Manufacturer: model.Manufacturer{
				Name:    "Boeing",
				Country: "USA",
				Website: "https://boeing.com",
			},
			Tags: []string{"wing", "delta", "atmospheric"},
			Metadata: map[string]interface{}{
				"max_temperature_c": 1650,
			},
			CreatedAt: now,
			UpdatedAt: &now,
		},
	}

	for _, p := range parts {
		r.parts[p.UUID] = p
	}
	return r
}
