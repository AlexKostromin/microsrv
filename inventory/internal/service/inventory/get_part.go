package inventory

import (
	"context"

	"github.com/AlexKostromin/microsrv/inventory/internal/model"
)

func (s *Service) GetPart(ctx context.Context, id string) (*model.Part, error) {
	return s.repo.GetPart(ctx, id)
}
