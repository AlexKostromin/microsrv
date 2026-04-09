package v1

import (
	"context"

	"github.com/AlexKostromin/microsrv/inventory/internal/converter"
	"github.com/AlexKostromin/microsrv/inventory/internal/model"
	inventoryV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/inventory/v1"
	"github.com/go-faster/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *Api) GetPart(ctx context.Context, req *inventoryV1.GetPartRequest) (*inventoryV1.GetPartResponse, error) {
	uuid := req.Uuid
	part, err := a.service.GetPart(ctx, uuid)

	if err != nil {
		if errors.Is(err, model.ErrPartNotFound) {
			return nil, status.Errorf(codes.NotFound, "part with UUID %s not found", uuid)
		}
		return nil, err
	}
	return &inventoryV1.GetPartResponse{
		Part: converter.ToPartFromService(part)}, nil
}
