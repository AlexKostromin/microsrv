package inventory

import (
	"github.com/AlexKostromin/microsrv/inventory/internal/model"
	repoMocks "github.com/AlexKostromin/microsrv/inventory/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"context"
	"testing"
)

func Test_Service_GetPart_Success(t *testing.T) {
	repo := repoMocks.NewInventoryRepository(t)
	svc := NewService(repo)
	expectedPart := &model.Part{
		UUID: "123", Name: "Raptor Engine", Description: "Raptor Engine",
	}
	repo.On("GetPart", mock.Anything, "123").Return(&model.Part{
		UUID: "123", Name: "Raptor Engine", Description: "Raptor Engine",
	}, nil)
	result, err := svc.GetPart(context.Background(), "123")
	assert.NoError(t, err)
	assert.Equal(t, expectedPart, result)
}
func Test_Service_GetPart_NotFound(t *testing.T) {
	repo := repoMocks.NewInventoryRepository(t)
	svc := NewService(repo)
	repo.On("GetPart", mock.Anything, "123").Return(nil, model.ErrPartNotFound)
	result, err := svc.GetPart(context.Background(), "123")
	assert.ErrorIs(t, err, model.ErrPartNotFound)
	assert.Nil(t, result)
}
