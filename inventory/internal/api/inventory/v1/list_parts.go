package v1

import (
	"context"

	"github.com/AlexKostromin/microsrv/inventory/internal/converter"
	"github.com/AlexKostromin/microsrv/inventory/internal/model"
	inventoryV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/inventory/v1"
)

func (a *Api) ListParts(ctx context.Context, req *inventoryV1.ListPartsRequest) (*inventoryV1.ListPartsResponse, error) {
	filter := req.Filter
	domainFilter := model.PartsFilter{
		UUIDS:                  filter.Uuids,
		Names:                  filter.Names,
		ManufacturersCountries: filter.ManufacturerCountries,
		Tags:                   filter.Tags,
	}
	categories := make([]model.Category, len(filter.Categories))
	for i, c := range filter.Categories {
		categories[i] = model.Category(c)
	}
	domainFilter.Categories = categories
	parts, err := a.service.ListParts(ctx, domainFilter)
	if err != nil {
		return nil, err
	}
	protoParts := make([]*inventoryV1.Part, len(parts))
	for i, part := range parts {
		protoParts[i] = converter.ToPartFromService(part)
	}
	return &inventoryV1.ListPartsResponse{
		Parts: protoParts}, nil
}
