package inventory

import (
	"context"

	"github.com/AlexKostromin/microsrv/inventory/internal/model"
	"github.com/AlexKostromin/microsrv/inventory/internal/repository/converter"
)

func (r *Repository) GetPart(ctx context.Context, id string) (*model.Part, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	part, ok := r.parts[id]
	if !ok {
		return nil, model.ErrPartNotFound
	}

	return converter.ToPartFromRepo(part), nil
}
