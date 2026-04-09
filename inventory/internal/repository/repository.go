package repository

import (
	"context"

	"github.com/AlexKostromin/microsrv/inventory/internal/model"
)

type InventoryRepository interface {
	GetPart(ctx context.Context, id string) (*model.Part, error)
	ListParts(ctx context.Context, filter model.PartsFilter) ([]*model.Part, error)
}
