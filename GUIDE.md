# Пошаговая инструкция: Неделя 2 — Слоистая архитектура и юнит-тесты

## Оглавление

1. [Обзор задания](#1-обзор-задания)
2. [Целевая структура проекта](#2-целевая-структура-проекта)
3. [Слоистая архитектура — принцип](#3-слоистая-архитектура--принцип)
4. [Доменные модели (model)](#4-доменные-модели-model)
5. [Конвертеры (converter)](#5-конвертеры-converter)
6. [Repository-слой](#6-repository-слой)
7. [Service-слой](#7-service-слой)
8. [API-слой](#8-api-слой)
9. [Точка входа (cmd/main.go)](#9-точка-входа-cmdmaingo)
10. [Моки и mockery](#10-моки-и-mockery)
11. [Юнит-тесты с testify/suite](#11-юнит-тесты-с-testifysuite)
12. [Применение к каждому сервису](#12-применение-к-каждому-сервису)
13. [Запуск и проверка](#13-запуск-и-проверка)

---

## 1. Обзор задания

На прошлой неделе ты реализовал три сервиса «всё в одном `main.go`». Теперь нужно:

1. **Рефакторинг** — разбить каждый сервис на слои: `api → service → repository`
2. **Доменные модели** — отделить бизнес-логику от proto/OpenAPI типов
3. **Конвертеры** — трансформация данных между слоями
4. **Юнит-тесты** — покрытие ≥ 80% через `testify/suite` + `mockery`

Три сервиса:
- **InventoryService** (gRPC, порт 50051) — `GetPart`, `ListParts` с фильтрацией
- **PaymentService** (gRPC, порт 50052) — `PayOrder`
- **OrderService** (HTTP, порт 8080) — `CreateOrder`, `GetOrder`, `PayOrder`, `CancelOrder`

---

## 2. Целевая структура проекта

Каждый сервис должен иметь одинаковую внутреннюю структуру. Пример для `inventory`:

```
inventory/
├── cmd/
│   └── main.go                        # точка входа, DI
├── internal/
│   ├── api/                           # gRPC/HTTP хендлеры
│   │   └── inventory/
│   │       └── v1/
│   │           ├── api.go             # структура + конструктор
│   │           ├── get_part.go        # обработчик GetPart
│   │           └── list_parts.go      # обработчик ListParts
│   ├── converter/                     # proto/openapi ↔ domain
│   │   └── part.go
│   ├── model/                         # доменные модели
│   │   ├── part.go
│   │   └── errors.go
│   ├── service/                       # интерфейс + реализация
│   │   ├── service.go                 # интерфейс InventoryService
│   │   └── inventory/
│   │       ├── service.go             # структура + конструктор
│   │       ├── get_part.go
│   │       └── list_parts.go
│   └── repository/                    # интерфейс + реализация
│       ├── repository.go              # интерфейс InventoryRepository
│       ├── converter/                 # domain ↔ repo model
│       │   └── part.go
│       ├── model/                     # модели хранилища
│       │   └── part.go
│       └── memory/                    # in-memory реализация
│           ├── repository.go          # структура + конструктор + seed
│           ├── get_part.go
│           └── list_parts.go
├── go.mod
└── go.sum
```

> **Принцип именования из курса:**
> - Файл с интерфейсом — `service.go` / `repository.go` в корне пакета
> - Реализация — в подпакете (`inventory/`, `memory/`)
> - Каждый метод — в отдельном файле (`get_part.go`, `list_parts.go`)

---

## 3. Слоистая архитектура — принцип

```
┌─────────────────────────┐
│  API-слой (хендлеры)    │  ← принимает proto/HTTP запросы
│  converter: proto→model │     конвертирует, вызывает service
├─────────────────────────┤
│  Service-слой           │  ← бизнес-логика
│  (бизнес-правила)       │     работает ТОЛЬКО с доменными моделями
├─────────────────────────┤
│  Repository-слой        │  ← доступ к данным
│  converter: model→repo  │     хранение в map / БД
└─────────────────────────┘
```

**Правило зависимостей:** каждый слой зависит ТОЛЬКО от интерфейса слоя ниже.
- API зависит от `service.InventoryService` (интерфейс)
- Service зависит от `repository.InventoryRepository` (интерфейс)
- Repository ни от чего не зависит

**Зачем?** Это позволяет:
- Тестировать каждый слой изолированно (подставляя моки)
- Менять реализацию хранилища (map → PostgreSQL) не трогая бизнес-логику
- Менять транспорт (gRPC → HTTP) не трогая сервисный слой

---

## 4. Доменные модели (model)

Доменная модель — это «чистые» Go-структуры, которые не знают ни про proto, ни про JSON, ни про базу данных.

**Пример для InventoryService (`internal/model/part.go`):**

```go
package model

import "time"

type Category int32

const (
    CategoryUnspecified Category = 0
    CategoryEngine      Category = 1
    CategoryFuel        Category = 2
    CategoryPorthole    Category = 3
    CategoryWing        Category = 4
)

type Dimensions struct {
    Length float64
    Width  float64
    Height float64
    Weight float64
}

type Manufacturer struct {
    Name    string
    Country string
    Website string
}

type Part struct {
    UUID          string
    Name          string
    Description   string
    Price         float64
    StockQuantity int64
    Category      Category
    Dimensions    Dimensions
    Manufacturer  Manufacturer
    Tags          []string
    Metadata      map[string]interface{}
    CreatedAt     time.Time
    UpdatedAt     *time.Time
}

type PartsFilter struct {
    UUIDs                 []string
    Names                 []string
    Categories            []Category
    ManufacturerCountries []string
    Tags                  []string
}
```

**Ошибки (`internal/model/errors.go`):**

```go
package model

import "errors"

var ErrPartNotFound = errors.New("part not found")
```

> **Зачем доменные ошибки?**
> API-слой ловит `model.ErrPartNotFound` и конвертирует его в `codes.NotFound` (gRPC)
> или `404` (HTTP). Сервисный и репо-слой не знают про HTTP-коды.

**Пример для OrderService (`internal/model/order.go`):**

```go
package model

type OrderStatus string

const (
    OrderStatusPendingPayment OrderStatus = "PENDING_PAYMENT"
    OrderStatusPaid           OrderStatus = "PAID"
    OrderStatusCancelled      OrderStatus = "CANCELLED"
)

type Order struct {
    UUID            string
    UserUUID        string
    PartUUIDs       []string
    TotalPrice      float64
    TransactionUUID *string
    PaymentMethod   *string
    Status          OrderStatus
}

var (
    ErrOrderNotFound    = errors.New("order not found")
    ErrOrderAlreadyPaid = errors.New("order already paid and cannot be cancelled")
)
```

---

## 5. Конвертеры (converter)

Конвертеры трансформируют данные между слоями. Каждый слой работает со своими типами.

### API-конвертер: proto ↔ domain

**Пример для InventoryService (`internal/converter/part.go`):**

```go
package converter

import (
    "google.golang.org/protobuf/types/known/timestamppb"

    "github.com/AlexKostromin/microsrv/inventory/internal/model"
    inventoryV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/inventory/v1"
)

// PartToProto конвертирует доменную модель в proto для ответа клиенту
func PartToProto(part model.Part) *inventoryV1.Part {
    protoPart := &inventoryV1.Part{
        Uuid:          part.UUID,
        Name:          part.Name,
        Description:   part.Description,
        Price:         part.Price,
        StockQuantity: part.StockQuantity,
        Category:      inventoryV1.Category(part.Category),
        Dimensions: &inventoryV1.Dimensions{
            Length: part.Dimensions.Length,
            Width:  part.Dimensions.Width,
            Height: part.Dimensions.Height,
            Weight: part.Dimensions.Weight,
        },
        Manufacturer: &inventoryV1.Manufacturer{
            Name:    part.Manufacturer.Name,
            Country: part.Manufacturer.Country,
            Website: part.Manufacturer.Website,
        },
        Tags:      part.Tags,
        CreatedAt: timestamppb.New(part.CreatedAt),
    }

    if part.UpdatedAt != nil {
        protoPart.UpdatedAt = timestamppb.New(*part.UpdatedAt)
    }

    return protoPart
}

// PartsFilterToModel конвертирует proto-фильтр в доменный
func PartsFilterToModel(filter *inventoryV1.PartsFilter) model.PartsFilter {
    if filter == nil {
        return model.PartsFilter{}
    }

    categories := make([]model.Category, 0, len(filter.Categories))
    for _, c := range filter.Categories {
        categories = append(categories, model.Category(c))
    }

    return model.PartsFilter{
        UUIDs:                 filter.Uuids,
        Names:                 filter.Names,
        Categories:            categories,
        ManufacturerCountries: filter.ManufacturerCountries,
        Tags:                  filter.Tags,
    }
}
```

> **Паттерн из курса:**
> - `XxxToProto()` — domain → proto (для ответов)
> - `XxxToModel()` — proto → domain (для запросов)
>
> Конвертер для каждого направления — отдельная функция.

### Конвертер для OrderService (OpenAPI ↔ domain)

OrderService использует ogen (не proto), поэтому конвертеры чуть другие:

```go
package converter

import (
    orderV1 "github.com/AlexKostromin/microsrv/shared/pkg/openapi/order/v1"
    "github.com/AlexKostromin/microsrv/order/internal/model"
)

func OrderToDTO(order model.Order) *orderV1.OrderDto {
    dto := &orderV1.OrderDto{
        UUID:       order.UUID,
        UserUUID:   order.UserUUID,
        PartUuids:  order.PartUUIDs,
        TotalPrice: order.TotalPrice,
        Status:     orderV1.OrderStatus(order.Status),
    }

    if order.TransactionUUID != nil {
        dto.TransactionUUID = orderV1.NewOptString(*order.TransactionUUID)
    }
    if order.PaymentMethod != nil {
        dto.PaymentMethod = orderV1.NewOptPaymentMethod(
            orderV1.PaymentMethod(*order.PaymentMethod),
        )
    }

    return dto
}
```

### Repo-конвертер: domain ↔ repo model

**Пример (`internal/repository/converter/part.go`):**

```go
package converter

import (
    "github.com/AlexKostromin/microsrv/inventory/internal/model"
    repoModel "github.com/AlexKostromin/microsrv/inventory/internal/repository/model"
)

func PartToRepoModel(part model.Part) repoModel.Part {
    return repoModel.Part{
        UUID:          part.UUID,
        Name:          part.Name,
        Description:   part.Description,
        Price:         part.Price,
        StockQuantity: part.StockQuantity,
        Category:      int32(part.Category),
        // ... остальные поля
    }
}

func PartToModel(part repoModel.Part) model.Part {
    return model.Part{
        UUID:          part.UUID,
        Name:          part.Name,
        Description:   part.Description,
        Price:         part.Price,
        StockQuantity: part.StockQuantity,
        Category:      model.Category(part.Category),
        // ... остальные поля
    }
}
```

> **Зачем два набора моделей (domain и repo)?**
> Сейчас они выглядят одинаково, но в будущем repo model будет иметь BSON-теги
> для MongoDB или SQL-теги для PostgreSQL, а domain model останется чистой.

---

## 6. Repository-слой

### Интерфейс

**`internal/repository/repository.go`:**

```go
package repository

import (
    "context"
    "github.com/AlexKostromin/microsrv/inventory/internal/model"
)

type InventoryRepository interface {
    GetPart(ctx context.Context, uuid string) (model.Part, error)
    ListParts(ctx context.Context, filter model.PartsFilter) ([]model.Part, error)
}
```

> **Правило из курса:** интерфейс лежит в корне пакета (`repository/repository.go`),
> а реализация — в подпакете (`repository/memory/`).

### Реализация (in-memory)

**`internal/repository/memory/repository.go`:**

```go
package memory

import (
    "sync"

    def "github.com/AlexKostromin/microsrv/inventory/internal/repository"
    repoModel "github.com/AlexKostromin/microsrv/inventory/internal/repository/model"
)

// Проверка на этапе компиляции, что repository реализует интерфейс
var _ def.InventoryRepository = (*repository)(nil)

type repository struct {
    mu   sync.RWMutex
    data map[string]repoModel.Part
}

func NewRepository() *repository {
    r := &repository{
        data: make(map[string]repoModel.Part),
    }
    r.seedData()
    return r
}
```

> **`var _ def.InventoryRepository = (*repository)(nil)`** — идиома Go.
> Если `repository` не реализует интерфейс, код не скомпилируется.
> Это ловит ошибки на этапе компиляции, а не в рантайме.

**`internal/repository/memory/get_part.go`:**

```go
package memory

import (
    "context"

    "github.com/AlexKostromin/microsrv/inventory/internal/model"
    repoConverter "github.com/AlexKostromin/microsrv/inventory/internal/repository/converter"
)

func (r *repository) GetPart(_ context.Context, uuid string) (model.Part, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    repoPart, ok := r.data[uuid]
    if !ok {
        return model.Part{}, model.ErrPartNotFound
    }

    return repoConverter.PartToModel(repoPart), nil
}
```

> **Обрати внимание:** репо возвращает `model.ErrPartNotFound` (доменную ошибку),
> а НЕ gRPC `status.Error`. Маппинг ошибок на транспорт — задача API-слоя.

**`internal/repository/memory/list_parts.go`** — фильтрация:

```go
func (r *repository) ListParts(_ context.Context, filter model.PartsFilter) ([]model.Part, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    result := make([]repoModel.Part, 0, len(r.data))
    for _, p := range r.data {
        result = append(result, p)
    }

    // Последовательная фильтрация: каждый шаг сужает результат
    if len(filter.UUIDs) > 0 {
        result = filterBy(result, func(p repoModel.Part) bool {
            return contains(filter.UUIDs, p.UUID)
        })
    }
    if len(filter.Names) > 0 {
        result = filterBy(result, func(p repoModel.Part) bool {
            return contains(filter.Names, p.Name)
        })
    }
    // ... аналогично для categories, countries, tags

    // Конвертируем в доменные модели
    parts := make([]model.Part, 0, len(result))
    for _, p := range result {
        parts = append(parts, repoConverter.PartToModel(p))
    }

    return parts, nil
}
```

> **Логика фильтрации (из контракта):**
> - Внутри одного поля — **ИЛИ** (uuid == "aaa" ИЛИ uuid == "bbb")
> - Между полями — **И** (uuid совпадает И категория совпадает)
> - Пустое поле = не фильтруем по нему

---

## 7. Service-слой

### Интерфейс

**`internal/service/service.go`:**

```go
package service

import (
    "context"
    "github.com/AlexKostromin/microsrv/inventory/internal/model"
)

type InventoryService interface {
    GetPart(ctx context.Context, uuid string) (model.Part, error)
    ListParts(ctx context.Context, filter model.PartsFilter) ([]model.Part, error)
}
```

### Реализация

**`internal/service/inventory/service.go`:**

```go
package inventory

import (
    "github.com/AlexKostromin/microsrv/inventory/internal/repository"
    def "github.com/AlexKostromin/microsrv/inventory/internal/service"
)

var _ def.InventoryService = (*service)(nil)

type service struct {
    inventoryRepository repository.InventoryRepository
}

func NewService(inventoryRepository repository.InventoryRepository) *service {
    return &service{
        inventoryRepository: inventoryRepository,
    }
}
```

**`internal/service/inventory/get_part.go`:**

```go
package inventory

import (
    "context"
    "github.com/AlexKostromin/microsrv/inventory/internal/model"
)

func (s *service) GetPart(ctx context.Context, uuid string) (model.Part, error) {
    part, err := s.inventoryRepository.GetPart(ctx, uuid)
    if err != nil {
        return model.Part{}, err
    }
    return part, nil
}
```

> **Сейчас сервис-слой простой** — почти pass-through к репозиторию.
> Но именно здесь будет бизнес-логика в будущих неделях:
> проверка прав, валидация, оркестрация нескольких вызовов.

### Service-слой OrderService — бизнес-логика

Для OrderService сервис-слой содержит реальную логику, потому что он оркестрирует
вызовы к Inventory и Payment. Сервис зависит от **двух** репозиториев
и **двух** внешних клиентов:

```go
package order

type service struct {
    orderRepository repository.OrderRepository
    inventoryClient client.InventoryClient   // gRPC-клиент
    paymentClient   client.PaymentClient     // gRPC-клиент
}
```

> **Где определять клиентские интерфейсы?**
> Создай пакет `internal/client/` с интерфейсами:
> ```go
> package client
>
> type InventoryClient interface {
>     ListParts(ctx context.Context, uuids []string) ([]model.Part, error)
> }
>
> type PaymentClient interface {
>     PayOrder(ctx context.Context, orderUUID, userUUID string, method model.PaymentMethod) (string, error)
> }
> ```
>
> Реализация — в `internal/client/inventory/` и `internal/client/payment/`,
> которые внутри вызывают настоящие gRPC-клиенты.

**Пример `internal/service/order/create_order.go`:**

```go
func (s *service) CreateOrder(ctx context.Context, userUUID string, partUUIDs []string) (model.Order, error) {
    // 1. Получаем детали из InventoryService
    parts, err := s.inventoryClient.ListParts(ctx, partUUIDs)
    if err != nil {
        return model.Order{}, fmt.Errorf("inventory error: %w", err)
    }

    // 2. Проверяем, что все детали найдены
    if len(parts) != len(partUUIDs) {
        return model.Order{}, fmt.Errorf("some parts not found")
    }

    // 3. Считаем цену
    var totalPrice float64
    for _, p := range parts {
        totalPrice += p.Price
    }

    // 4. Создаём заказ
    order := model.Order{
        UUID:       uuid.NewString(),
        UserUUID:   userUUID,
        PartUUIDs:  partUUIDs,
        TotalPrice: totalPrice,
        Status:     model.OrderStatusPendingPayment,
    }

    err = s.orderRepository.Create(ctx, order)
    if err != nil {
        return model.Order{}, err
    }

    return order, nil
}
```

---

## 8. API-слой

### gRPC API (InventoryService)

**`internal/api/inventory/v1/api.go`:**

```go
package v1

import (
    "github.com/AlexKostromin/microsrv/inventory/internal/service"
    inventoryV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/inventory/v1"
)

type api struct {
    inventoryV1.UnimplementedInventoryServiceServer
    inventoryService service.InventoryService
}

func NewAPI(inventoryService service.InventoryService) *api {
    return &api{
        inventoryService: inventoryService,
    }
}
```

**`internal/api/inventory/v1/get_part.go`:**

```go
package v1

import (
    "context"
    "errors"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    "github.com/AlexKostromin/microsrv/inventory/internal/converter"
    "github.com/AlexKostromin/microsrv/inventory/internal/model"
    inventoryV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/inventory/v1"
)

func (a *api) GetPart(ctx context.Context, req *inventoryV1.GetPartRequest) (*inventoryV1.GetPartResponse, error) {
    part, err := a.inventoryService.GetPart(ctx, req.GetUuid())
    if err != nil {
        // Маппинг доменной ошибки → gRPC статус
        if errors.Is(err, model.ErrPartNotFound) {
            return nil, status.Errorf(codes.NotFound, "part with UUID %s not found", req.GetUuid())
        }
        return nil, err
    }

    return &inventoryV1.GetPartResponse{
        Part: converter.PartToProto(part),
    }, nil
}
```

> **Ключевой паттерн:** API-слой отвечает за:
> 1. Конвертацию proto → domain (через `converter`)
> 2. Вызов service-метода
> 3. Маппинг доменных ошибок в gRPC/HTTP ответы
> 4. Конвертацию domain → proto для ответа

### HTTP API (OrderService)

Для OrderService с ogen API-слой реализует интерфейс `orderV1.Handler`:

```go
package v1

import (
    "github.com/AlexKostromin/microsrv/order/internal/service"
    orderV1 "github.com/AlexKostromin/microsrv/shared/pkg/openapi/order/v1"
)

type api struct {
    orderService service.OrderService
}

func NewAPI(orderService service.OrderService) *api {
    return &api{orderService: orderService}
}
```

**Обработка ошибок в HTTP API (ogen):**

```go
func (a *api) GetOrder(ctx context.Context, params orderV1.GetOrderParams) (orderV1.GetOrderRes, error) {
    order, err := a.orderService.GetOrder(ctx, params.OrderUUID.String())
    if err != nil {
        if errors.Is(err, model.ErrOrderNotFound) {
            return &orderV1.NotFoundError{Code: 404, Message: err.Error()}, nil
        }
        return &orderV1.InternalServerError{Code: 500, Message: err.Error()}, nil
    }

    return converter.OrderToDTO(order), nil
}
```

> **Разница с gRPC:** в ogen ошибки возвращаются как `(СтруктураОшибки, nil)`,
> а не `(nil, error)`. Ogen сам определяет HTTP-код по типу возвращённой структуры:
> - `*NotFoundError` → **404**
> - `*ConflictError` → **409**
> - `*CancelOrderNoContent{}` → **204**

---

## 9. Точка входа (cmd/main.go)

Здесь происходит **Dependency Injection** — ручная сборка зависимостей:

```go
func main() {
    // 1. Собираем зависимости снизу вверх
    repo := memoryRepo.NewRepository()        // Repository
    svc := inventorySvc.NewService(repo)       // Service(Repository)
    apiHandler := inventoryAPI.NewAPI(svc)     // API(Service)

    // 2. Создаём gRPC-сервер
    s := grpc.NewServer()
    inventoryV1.RegisterInventoryServiceServer(s, apiHandler)
    reflection.Register(s)

    // 3. Запускаем
    lis, _ := net.Listen("tcp", ":50051")
    go s.Serve(lis)

    // 4. Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    s.GracefulStop()
}
```

> **Порядок: снизу вверх.**
> Сначала создаёшь репозиторий, потом сервис (передаёшь репо), потом API (передаёшь сервис).
> Каждый конструктор принимает **интерфейс**, а не конкретную реализацию.

Для **OrderService** в main.go дополнительно создаются gRPC-клиенты:

```go
// gRPC-клиенты к другим сервисам
invConn, _ := grpc.NewClient("localhost:50051",
    grpc.WithTransportCredentials(insecure.NewCredentials()))
invClient := inventoryClient.NewClient(invConn)

payConn, _ := grpc.NewClient("localhost:50052",
    grpc.WithTransportCredentials(insecure.NewCredentials()))
payClient := paymentClient.NewClient(payConn)

// DI
repo := memoryRepo.NewRepository()
svc := orderSvc.NewService(repo, invClient, payClient)
apiHandler := orderAPI.NewAPI(svc)
```

---

## 10. Моки и mockery

Моки нужны для тестирования каждого слоя изолированно:
- Тесты **API-слоя** — мокаем Service
- Тесты **Service-слоя** — мокаем Repository (и Client)

### Установка mockery

```bash
go install github.com/vektra/mockery/v2@latest
```

### Конфигурация

Создай **`.mockery.yaml`** в корне сервиса:

```yaml
with-expecter: true
dir: "{{.InterfaceDir}}/mocks"
filename: "mock_{{.InterfaceNameSnake}}.go"
outpkg: "mocks"
mockname: "{{.InterfaceName}}"

packages:
  github.com/AlexKostromin/microsrv/inventory/internal/service:
    config:
      include-regex: ".*Service"

  github.com/AlexKostromin/microsrv/inventory/internal/repository:
    config:
      include-regex: ".*Repository"
```

### Генерация моков

```bash
cd inventory && mockery
```

Это создаст:
- `internal/service/mocks/mock_inventory_service.go`
- `internal/repository/mocks/mock_inventory_repository.go`

> **Что генерирует mockery?**
> Для каждого метода интерфейса — мок-реализацию, которая:
> - Записывает вызовы (аргументы, количество)
> - Возвращает заранее настроенные значения
> - Проверяет, что все ожидания выполнены

---

## 11. Юнит-тесты с testify/suite

### Что такое Suite

`testify/suite` — фреймворк, который даёт `SetupTest()` (вызывается перед КАЖДЫМ тестом)
и `TearDownTest()` (после каждого). Это гарантирует, что каждый тест начинает с чистого состояния.

### Suite для API-слоя

**`internal/api/inventory/v1/suite_test.go`:**

```go
package v1

import (
    "context"
    "testing"

    "github.com/stretchr/testify/suite"

    "github.com/AlexKostromin/microsrv/inventory/internal/service/mocks"
)

type APISuite struct {
    suite.Suite

    ctx              context.Context
    inventoryService *mocks.InventoryService
    api              *api
}

func (s *APISuite) SetupTest() {
    s.ctx = context.Background()
    s.inventoryService = mocks.NewInventoryService(s.T())
    s.api = NewAPI(s.inventoryService)
}

func TestAPISuite(t *testing.T) {
    suite.Run(t, new(APISuite))
}
```

> **`mocks.NewInventoryService(s.T())`** — создаёт мок, привязанный к текущему тесту.
> Если ожидаемый вызов не произошёл, тест автоматически упадёт при cleanup.

### Тесты методов

**`internal/api/inventory/v1/get_part_test.go`:**

```go
package v1

import (
    "github.com/brianvoe/gofakeit/v7"

    "github.com/AlexKostromin/microsrv/inventory/internal/converter"
    "github.com/AlexKostromin/microsrv/inventory/internal/model"
    inventoryV1 "github.com/AlexKostromin/microsrv/shared/pkg/proto/inventory/v1"
)

func (s *APISuite) TestGetPartSuccess() {
    partUUID := gofakeit.UUID()

    expectedPart := model.Part{
        UUID:  partUUID,
        Name:  gofakeit.ProductName(),
        Price: gofakeit.Price(100, 10000),
    }

    // Настраиваем мок: при вызове GetPart с этим UUID — вернуть expectedPart
    s.inventoryService.On("GetPart", s.ctx, partUUID).Return(expectedPart, nil)

    req := &inventoryV1.GetPartRequest{Uuid: partUUID}
    res, err := s.api.GetPart(s.ctx, req)

    s.Require().NoError(err)
    s.Require().NotNil(res)
    s.Require().Equal(partUUID, res.GetPart().GetUuid())
}

func (s *APISuite) TestGetPartNotFound() {
    partUUID := gofakeit.UUID()

    s.inventoryService.On("GetPart", s.ctx, partUUID).
        Return(model.Part{}, model.ErrPartNotFound)

    req := &inventoryV1.GetPartRequest{Uuid: partUUID}
    res, err := s.api.GetPart(s.ctx, req)

    s.Require().Error(err)
    s.Require().Nil(res)
    // Проверяем, что вернулся именно NotFound
    s.Require().Contains(err.Error(), "not found")
}
```

> **Паттерн тестирования (из курса):**
> 1. Настрой мок: `s.serviceMock.On("Method", args...).Return(results...)`
> 2. Вызови тестируемый метод
> 3. Проверь результат через `s.Require().NoError()`, `s.Require().Equal()`
>
> **`s.Require()`** — fail-fast: если проверка не прошла, тест останавливается.
> **`s.Assert()`** — soft: тест продолжается, но в конце считается упавшим.

### Suite для Service-слоя

**`internal/service/inventory/suite_test.go`:**

```go
package inventory

import (
    "context"
    "testing"

    "github.com/stretchr/testify/suite"

    "github.com/AlexKostromin/microsrv/inventory/internal/repository/mocks"
)

type ServiceSuite struct {
    suite.Suite

    ctx                 context.Context
    inventoryRepository *mocks.InventoryRepository
    service             *service
}

func (s *ServiceSuite) SetupTest() {
    s.ctx = context.Background()
    s.inventoryRepository = mocks.NewInventoryRepository(s.T())
    s.service = NewService(s.inventoryRepository)
}

func TestServiceSuite(t *testing.T) {
    suite.Run(t, new(ServiceSuite))
}
```

**`internal/service/inventory/get_part_test.go`:**

```go
func (s *ServiceSuite) TestGetPartSuccess() {
    partUUID := gofakeit.UUID()
    expectedPart := model.Part{UUID: partUUID, Name: gofakeit.ProductName()}

    s.inventoryRepository.On("GetPart", s.ctx, partUUID).Return(expectedPart, nil)

    part, err := s.service.GetPart(s.ctx, partUUID)

    s.Require().NoError(err)
    s.Require().Equal(expectedPart, part)
}

func (s *ServiceSuite) TestGetPartRepoError() {
    repoErr := gofakeit.Error()
    partUUID := gofakeit.UUID()

    s.inventoryRepository.On("GetPart", s.ctx, partUUID).Return(model.Part{}, repoErr)

    part, err := s.service.GetPart(s.ctx, partUUID)

    s.Require().Error(err)
    s.Require().ErrorIs(err, repoErr)
    s.Require().Empty(part.UUID)
}
```

> **Для каждого метода минимум 2 теста:**
> 1. Успешный сценарий
> 2. Ошибка от нижнего слоя

### gofakeit — генерация тестовых данных

```go
import "github.com/brianvoe/gofakeit/v7"

uuid := gofakeit.UUID()          // случайный UUID
name := gofakeit.ProductName()   // "Incredible Granite Chair"
price := gofakeit.Price(1, 9999) // 4523.78
city := gofakeit.City()          // "North Alana"
err := gofakeit.Error()          // случайная ошибка
```

> **Зачем случайные данные?** Чтобы тесты не зависели от конкретных значений.
> Если тест падает только с определёнными данными — это баг в коде.

---

## 12. Применение к каждому сервису

### InventoryService

| Слой | Интерфейс | Методы |
|------|-----------|--------|
| API | `InventoryServiceServer` (proto) | `GetPart`, `ListParts` |
| Service | `InventoryService` | `GetPart`, `ListParts` |
| Repository | `InventoryRepository` | `GetPart`, `ListParts` |

- Конвертеры: `proto ↔ model` + `model ↔ repoModel`
- Seed data: 3-4 детали разных категорий (ENGINE, FUEL, PORTHOLE, WING)
- Фильтрация: реализуется в repository-слое

### PaymentService

| Слой | Интерфейс | Методы |
|------|-----------|--------|
| API | `PaymentServiceServer` (proto) | `PayOrder` |
| Service | `PaymentService` | `PayOrder` |

- Самый простой сервис — нет репозитория
- Service генерирует `transaction_uuid` и логирует в консоль
- Конвертеры минимальные

### OrderService

| Слой | Интерфейс | Методы |
|------|-----------|--------|
| API | `orderV1.Handler` (ogen) | `CreateOrder`, `GetOrder`, `PayOrder`, `CancelOrder`, `NewError` |
| Service | `OrderService` | `CreateOrder`, `GetOrder`, `PayOrder`, `CancelOrder` |
| Repository | `OrderRepository` | `Create`, `Get`, `UpdateStatus` |
| Client | `InventoryClient` | `ListParts` |
| Client | `PaymentClient` | `PayOrder` |

- Бизнес-логика в service-слое: проверка деталей, расчёт цены, проверка статуса при отмене
- API-слой маппит доменные ошибки на HTTP-коды через типы ogen
- `internal/client/` — обёртки над gRPC-клиентами с конвертацией proto → domain

---

## 13. Запуск и проверка

### Запуск тестов

```bash
# Все тесты с покрытием
go test ./inventory/... -cover
go test ./payment/... -cover
go test ./order/... -cover

# Тесты конкретного слоя
go test ./inventory/internal/api/... -v
go test ./inventory/internal/service/... -v
```

### Запуск сервисов

```bash
# Терминал 1
go run ./inventory/cmd/...

# Терминал 2
go run ./payment/cmd/...

# Терминал 3
go run ./order/cmd/...
```

### Автотесты API

```bash
task test-api
```

---

## Частые ошибки

| Проблема | Причина | Решение |
|----------|---------|---------|
| `cannot use *service as Service` | Метод не совпадает с интерфейсом | Добавь `var _ def.Service = (*service)(nil)` для раннего обнаружения |
| Мок не вызван → тест падает | Настроил `.On()` но метод не вызвался | Проверь аргументы — мок сравнивает точно |
| `import cycle` | Слой импортирует вышестоящий | Используй интерфейсы, а не конкретные типы |
| Тесты влияют друг на друга | Общее состояние между тестами | `SetupTest()` создаёт новые моки для каждого теста |
| `mockery` не генерирует | Неправильные пути в `.mockery.yaml` | Путь пакета должен совпадать с `module` в `go.mod` |

---

## Чеклист перед сдачей

- [ ] Каждый сервис разбит на `api / service / repository` слои
- [ ] Каждый слой зависит только от **интерфейса** нижнего слоя
- [ ] Доменные модели отделены от proto/OpenAPI типов
- [ ] Конвертеры: proto ↔ domain ↔ repo model
- [ ] Моки сгенерированы через mockery
- [ ] Юнит-тесты API-слоя: мокаем service
- [ ] Юнит-тесты Service-слоя: мокаем repository
- [ ] `go test ./... -cover` показывает ≥ 80%
- [ ] `task test-api` проходит (все 9 тестов)
