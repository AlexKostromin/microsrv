package converter

import (
	"testing"
	"time"

	"github.com/AlexKostromin/microsrv/inventory/internal/model"
	"github.com/stretchr/testify/assert"
)

func Test_ToPartFromService_Success(t *testing.T) {
	now := time.Now()

	input := &model.Part{
		UUID:          "550e8400-e29b-41d4-a716-446655440000",
		Name:          "Raptor Engine V2",
		Description:   "Основной двигатель для межпланетных перелётов",
		Price:         1500000.00,
		StockQuantity: 10,
		Category:      model.CategoryEngine,
		Dimensions: model.Dimensions{
			Length: 380,
			Width:  130,
			Height: 130,
			Weight: 1600,
		},
		Manufacturer: model.Manufacturer{
			Name:    "SpaceX",
			Country: "USA",
			Website: "https://spacex.com",
		},
		Tags: []string{"engine", "raptor"},
		Metadata: map[string]interface{}{
			"label":    "test-value",
			"thrust":   3.14,
			"reusable": true,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := ToPartFromService(input)

	assert.NotNil(t, result)
	assert.Equal(t, input.UUID, result.Uuid)
	assert.Equal(t, input.Name, result.Name)
	assert.Equal(t, input.Description, result.Description)
	assert.Equal(t, input.Price, result.Price)
	assert.Equal(t, input.StockQuantity, result.StockQuantity)
	assert.Equal(t, input.Dimensions.Length, result.Dimensions.Length)
	assert.Equal(t, input.Manufacturer.Name, result.Manufacturer.Name)
	assert.Equal(t, input.Tags, result.Tags)
	assert.Equal(t, "test-value", result.Metadata["label"].GetStringValue())
	assert.Equal(t, 3.14, result.Metadata["thrust"].GetDoubleValue())
	assert.Equal(t, true, result.Metadata["reusable"].GetBoolValue())
}

func Test_ToPartFromService_EmptyMetadata(t *testing.T) {
	input := &model.Part{
		UUID: "550e8400-e29b-41d4-a716-446655440001",
		Name: "Test Part",
	}

	result := ToPartFromService(input)

	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result.Metadata))
}
