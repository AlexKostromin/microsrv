package inventory

import (
	"github.com/AlexKostromin/microsrv/inventory/internal/repository"
	"github.com/AlexKostromin/microsrv/inventory/internal/service"
)

type Service struct {
	repo repository.InventoryRepository
}

var _ service.InventoryService = (*Service)(nil)

func NewService(repo repository.InventoryRepository) *Service {
	return &Service{
		repo: repo,
	}
}
