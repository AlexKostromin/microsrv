package inventory

import (
	"context"
	"testing"

	"github.com/AlexKostromin/microsrv/inventory/internal/model"
	"github.com/stretchr/testify/assert"
)

func Test_Repository_GetPart_Success(t *testing.T) {
	// 1. Создай репозиторий
	repo := New()
	// 2. Достань любой UUID из repo.parts (это map — как пройтись по map в Go?)
	var UUID string

	for _, v := range repo.parts {
		UUID = v.UUID
		break
	}

	// 3. Вызови repo.GetPart(context.Background(), uuid)
	result, err := repo.GetPart(context.Background(), UUID)
	// 4. Проверь: ошибка == nil? Результат != nil?
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func Test_Repository_GetPart_NotFound(t *testing.T) {
	// 1. Создай репозиторий
	repo := New()

	// 2. Вызови repo.GetPart с несуществующим UUID (например "non-existent")
	result, err := repo.GetPart(context.Background(), "412231")
	// 3. Проверь: ошибка == model.ErrPartNotFound?
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, model.ErrPartNotFound)
	assert.Nil(t, result)
}
