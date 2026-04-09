package inventory

import (
	"context"

	"github.com/AlexKostromin/microsrv/inventory/internal/model"
	"github.com/AlexKostromin/microsrv/inventory/internal/repository/converter"
	repoModel "github.com/AlexKostromin/microsrv/inventory/internal/repository/model"
)

func (r *Repository) ListParts(ctx context.Context, filter model.PartsFilter) ([]*model.Part, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Собираем все детали
	result := make([]*repoModel.Part, 0, len(r.parts))
	for _, p := range r.parts {
		result = append(result, p)
	}

	if len(filter.UUIDS) > 0 {
		uuidSet := make(map[string]struct{}, len(filter.UUIDS))
		for _, u := range filter.UUIDS {
			uuidSet[u] = struct{}{}
		}
		result = filterParts(result, func(p *repoModel.Part) bool {
			_, ok := uuidSet[p.UUID]
			return ok
		})
	}

	// Фильтрация по имени
	if len(filter.Names) > 0 {
		nameSet := make(map[string]struct{}, len(filter.Names))
		for _, n := range filter.Names {
			nameSet[n] = struct{}{}
		}
		result = filterParts(result, func(p *repoModel.Part) bool {
			_, ok := nameSet[p.Name]
			return ok
		})
	}

	// Фильтрация по категории
	if len(filter.Categories) > 0 {
		catSet := make(map[model.Category]struct{}, len(filter.Categories))
		for _, c := range filter.Categories {
			catSet[c] = struct{}{}
		}
		result = filterParts(result, func(p *repoModel.Part) bool {
			_, ok := catSet[model.Category(p.Category)]
			return ok
		})
	}

	// Фильтрация по стране производителя
	if len(filter.ManufacturersCountries) > 0 {
		countrySet := make(map[string]struct{}, len(filter.ManufacturersCountries))
		for _, c := range filter.ManufacturersCountries {
			countrySet[c] = struct{}{}
		}
		result = filterParts(result, func(p *repoModel.Part) bool {
			_, ok := countrySet[p.Manufacturer.Country]
			return ok
		})
	}

	// Фильтрация по тегам
	if len(filter.Tags) > 0 {
		tagSet := make(map[string]struct{}, len(filter.Tags))
		for _, t := range filter.Tags {
			tagSet[t] = struct{}{}
		}
		result = filterParts(result, func(p *repoModel.Part) bool {
			for _, tag := range p.Tags {
				if _, ok := tagSet[tag]; ok {
					return true
				}
			}
			return false
		})
	}
	domainResult := make([]*model.Part, 0, len(result))
	for _, p := range result {
		domainResult = append(domainResult, converter.ToPartFromRepo(p))
	}
	return domainResult, nil
}

func filterParts(parts []*repoModel.Part, predicate func(*repoModel.Part) bool) []*repoModel.Part {
	filtered := make([]*repoModel.Part, 0)
	for _, p := range parts {
		if predicate(p) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
