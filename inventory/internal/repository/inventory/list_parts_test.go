package inventory

import (
	"context"
	"testing"

	"github.com/AlexKostromin/microsrv/inventory/internal/model"
	"github.com/stretchr/testify/assert"
)

func Test_Repository_ListParts_EmptyFilter(t *testing.T) {
	// 1. Создай репозиторий через New()
	repo := New()
	// 2. Создай пустой фильтр: model.PartsFilter{}
	filter := model.PartsFilter{}
	// 3. Вызови repo.ListParts(context.Background(), filter)
	result, err := repo.ListParts(context.Background(), filter)
	// 4. Проверь: ошибка == nil?
	assert.NoError(t, err)
	// 5. Проверь: len(result) == 4? (в репозитории 4 детали по умолчанию)
	assert.Equal(t, 4, len(result))
}

func Test_Repository_ListParts_ByName(t *testing.T) {
	// 1. Создай репозиторий через New()
	repo := New()
	// 2. Создай фильтр с именем: model.PartsFilter{Names: []string{"Raptor Engine V2"}}
	filter := model.PartsFilter{
		Names: []string{"Raptor Engine V2"},
	}
	// 3. Вызови repo.ListParts(context.Background(), filter)
	result, err := repo.ListParts(context.Background(), filter)
	// 4. Проверь: ошибка == nil?
	assert.NoError(t, err)
	// 5. Проверь: len(result) == 1?
	assert.Equal(t, 1, len(result))
	// 6. Проверь: result[0].Name == "Raptor Engine V2"?
	assert.Equal(t, "Raptor Engine V2", result[0].Name)
}

func Test_Repository_ListParts_ByCategory(t *testing.T) {
	// 1. Создай репозиторий через New()
	repo := New()
	// 2. Создай фильтр с категорией: model.PartsFilter{Categories: []model.Category{model.CategoryEngine}}
	filter := model.PartsFilter{
		Categories: []model.Category{
			model.CategoryEngine,
		},
	}
	// 3. Вызови repo.ListParts(context.Background(), filter)
	result, err := repo.ListParts(context.Background(), filter)
	// 4. Проверь: ошибка == nil?
	assert.NoError(t, err)
	// 5. Проверь: len(result) == 1? (только Raptor Engine V2 — категория Engine)
	assert.Equal(t, 1, len(result))
}

func Test_Repository_ListParts_ByCountry(t *testing.T) {
	// 1. Создай репозиторий через New()
	repo := New()
	// 2. Создай фильтр по стране: model.PartsFilter{ManufacturersCountries: []string{"USA"}}
	filter := model.PartsFilter{
		ManufacturersCountries: []string{"USA"},
	}
	// 3. Вызови repo.ListParts(context.Background(), filter)
	result, err := repo.ListParts(context.Background(), filter)
	// 4. Проверь: ошибка == nil?
	assert.NoError(t, err)
	// 5. Проверь: len(result) == 2? (SpaceX — USA и Boeing — USA)
	assert.Equal(t, 2, len(result))
}

func Test_Repository_ListParts_ByTag(t *testing.T) {
	// 1. Создай репозиторий через New()
	repo := New()
	// 2. Создай фильтр по тегу: model.PartsFilter{Tags: []string{"methane"}}
	filter := model.PartsFilter{
		Tags: []string{"methane"},
	}
	// 3. Вызови repo.ListParts(context.Background(), filter)
	result, err := repo.ListParts(context.Background(), filter)
	// 4. Проверь: ошибка == nil?
	assert.NoError(t, err)
	// 5. Проверь: len(result) == 2? (Raptor Engine V2 и Liquid Methane Tank — оба имеют тег "methane")
	assert.Equal(t, 2, len(result))
}

func Test_Repository_ListParts_ByUUID(t *testing.T) {
	// 1. Создай репозиторий через New()
	repo := New()
	// 2. Достань любой UUID из repo.parts (пройдись по map, возьми первый — как в get_part_test.go)
	var UUID string
	for _, v := range repo.parts {
		UUID = v.UUID
		break
	}
	// 3. Создай фильтр: model.PartsFilter{UUIDS: []string{uuid}}
	filter := model.PartsFilter{
		UUIDS: []string{UUID},
	}
	// 4. Вызови repo.ListParts(context.Background(), filter)
	result, err := repo.ListParts(context.Background(), filter)
	// 5. Проверь: ошибка == nil?
	assert.NoError(t, err)
	// 6. Проверь: len(result) == 1?
	assert.Equal(t, 1, len(result))
	// 7. Проверь: result[0].UUID == uuid?
	assert.Equal(t, UUID, result[0].UUID)
}
