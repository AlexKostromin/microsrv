package inventory

import (
	"context"
	"testing"

	"github.com/AlexKostromin/microsrv/inventory/internal/model"
	repoMocks "github.com/AlexKostromin/microsrv/inventory/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Repository_ListParts_Success(t *testing.T) {
	repo := repoMocks.NewInventoryRepository(t)
	svc := NewService(repo)
	filter := model.PartsFilter{
		UUIDS:                  []string{"123"},
		Names:                  []string{"Raptor Engine"},
		Tags:                   []string{"Raptor Engine"},
		ManufacturersCountries: []string{"Raptor Engine"},
		Categories:             make([]model.Category, 0),
	}
	expectedPart := []*model.Part{
		{UUID: "123", Name: "Raptor Engine"},
		{UUID: "321", Name: "Raptor Engine1"},
	}
	repo.On("ListParts", mock.Anything, filter).Return([]*model.Part{
		{UUID: "123", Name: "Raptor Engine"},
		{UUID: "321", Name: "Raptor Engine1"},
	}, nil)
	result, err := svc.ListParts(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, expectedPart, result)
}
func Test_Repository_ListParts_NotFound(t *testing.T) {
	repo := repoMocks.NewInventoryRepository(t)
	filter := model.PartsFilter{
		UUIDS:                  []string{"123"},
		Names:                  []string{"Raptor Engine"},
		Tags:                   []string{"Raptor Engine"},
		ManufacturersCountries: []string{"Raptor Engine"},
		Categories:             make([]model.Category, 0),
	}
	svc := NewService(repo)

	repo.On("ListParts", mock.Anything, filter).Return(nil, model.ErrPartNotFound)
	result, err := svc.ListParts(context.Background(), filter)
	assert.Error(t, err)
	assert.Nil(t, result)
}
