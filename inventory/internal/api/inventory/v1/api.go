package v1

import (
	"github.com/AlexKostromin/microsrv/inventory/internal/service"
	inventoryV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/inventory/v1"
)

type Api struct {
	service service.InventoryService
	inventoryV1.UnimplementedInventoryServiceServer
}

func NewApi(service service.InventoryService) *Api {
	return &Api{
		service: service,
	}
}
