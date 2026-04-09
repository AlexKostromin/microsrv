package converter

import (
	"testing"
	"time"

	"github.com/AlexKostromin/microsrv/inventory/internal/model"
	repoModel "github.com/AlexKostromin/microsrv/inventory/internal/repository/model"
	"github.com/stretchr/testify/assert"
)

func Test_ToPartFromRepo_Success(t *testing.T) {
	now := time.Now()

	input := &repoModel.Part{
		UUID:          "550e8400-e29b-41d4-a716-446655440000",
		Name:          "Raptor Engine V2",
		Description:   "Основной двигатель для межпланетных перелётов",
		Price:         1500000.00,
		StockQuantity: 10,
		Category:      1,
		Dimensions: repoModel.Dimensions{
			Length: 380,
			Width:  130,
			Height: 130,
			Weight: 1600,
		},
		Manufacturer: repoModel.Manufacturer{
			Name:    "SpaceX",
			Country: "USA",
			Website: "https://spacex.com",
		},
		Tags:      []string{"engine", "raptor"},
		Metadata:  map[string]interface{}{"thrust_kn": 2300},
		CreatedAt: now,
		UpdatedAt: &now,
	}

	result := ToPartFromRepo(input)

	assert.NotNil(t, result)
	assert.Equal(t, input.UUID, result.UUID)
	assert.Equal(t, input.Name, result.Name)
	assert.Equal(t, input.Description, result.Description)
	assert.Equal(t, input.Price, result.Price)
	assert.Equal(t, input.StockQuantity, result.StockQuantity)
	assert.Equal(t, model.Category(input.Category), result.Category)
	assert.Equal(t, input.Dimensions.Length, result.Dimensions.Length)
	assert.Equal(t, input.Manufacturer.Name, result.Manufacturer.Name)
	assert.Equal(t, input.Tags, result.Tags)
	assert.Equal(t, now, result.UpdatedAt)
}

func Test_ToPartFromRepo_NilUpdatedAt(t *testing.T) {
	input := &repoModel.Part{
		UUID:      "550e8400-e29b-41d4-a716-446655440001",
		Name:      "Test Part",
		UpdatedAt: nil,
	}

	result := ToPartFromRepo(input)

	assert.NotNil(t, result)
	assert.Equal(t, time.Time{}, result.UpdatedAt)
}
