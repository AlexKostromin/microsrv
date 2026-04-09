package converter

import (
	"github.com/AlexKostromin/microsrv/inventory/internal/model"
	repoModel "github.com/AlexKostromin/microsrv/inventory/internal/repository/model"
)

func ToPartFromRepo(part *repoModel.Part) *model.Part {
	p := model.Part{
		UUID:          part.UUID,
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		StockQuantity: part.StockQuantity,
		Category:      model.Category(part.Category),
		Dimensions:    model.Dimensions(part.Dimensions),
		Manufacturer:  model.Manufacturer(part.Manufacturer),
		Tags:          part.Tags,
		Metadata:      part.Metadata,
		CreatedAt:     part.CreatedAt,
	}
	if part.UpdatedAt != nil {
		p.UpdatedAt = *part.UpdatedAt
	}
	return &p
}
