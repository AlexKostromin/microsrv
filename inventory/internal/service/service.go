package service

import (
	"context"

	"github.com/AlexKostromin/microsrv/inventory/internal/model"
)

type InventoryService interface {
	GetPart(ctx context.Context, id string) (*model.Part, error)
	ListParts(ctx context.Context, filter model.PartsFilter) ([]*model.Part, error)
}
