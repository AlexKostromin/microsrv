package converter

import (
	"github.com/AlexKostromin/microsrv/inventory/internal/model"
	inventoryV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/inventory/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToPartFromService(part *model.Part) *inventoryV1.Part {

	p := inventoryV1.Part{
		Uuid:          part.UUID,
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		StockQuantity: part.StockQuantity,
		Category:      inventoryV1.Category(part.Category),
		Dimensions:    &inventoryV1.Dimensions{Length: part.Dimensions.Length, Width: part.Dimensions.Width, Height: part.Dimensions.Height, Weight: part.Dimensions.Weight},
		Manufacturer:  &inventoryV1.Manufacturer{Name: part.Manufacturer.Name, Country: part.Manufacturer.Country, Website: part.Manufacturer.Website},
		Tags:          part.Tags,
		Metadata:      make(map[string]*inventoryV1.Value),
		CreatedAt:     timestamppb.New(part.CreatedAt),
		UpdatedAt:     timestamppb.New(part.UpdatedAt),
	}
	for key, value := range part.Metadata {
		switch v := value.(type) {
		case string:
			p.Metadata[key] = &inventoryV1.Value{Kind: &inventoryV1.Value_StringValue{StringValue: v}}
		case int64:
			p.Metadata[key] = &inventoryV1.Value{Kind: &inventoryV1.Value_Int64Value{Int64Value: v}}
		case float64:
			p.Metadata[key] = &inventoryV1.Value{Kind: &inventoryV1.Value_DoubleValue{DoubleValue: v}}
		case bool:
			p.Metadata[key] = &inventoryV1.Value{Kind: &inventoryV1.Value_BoolValue{BoolValue: v}}
		}
	}

	return &p
}
