---
apply: always
---

# Общая архитектура и правила

## Слои и зависимости
- Transport слой: `internal/handler/grpc/*`.
  - Должен работать только с protobuf DTO и usecase-интерфейсами.
  - Не обращаться напрямую к репозиториям и сервисам.
- Usecase слой: `internal/usecase/*`.
  - Это входной слой от транспортного слоя (запросы от внешних систем)
  - Валидация входных параметров
  - orchestration доменных сервисов и сервисов `internal/service/*`.
  - Вход/выход — доменные модели (не protobuf).
  - желательно не обращается в соседние usecases
- Domain слой: `internal/domain/*`.
  - `model/` — структуры данных, сущности (entity).
  - `service/` — доменные операции и инварианты. Может использовать только репозиорий.
  - `repo/` — доступ к хранилищам. Может содержать подпапки для разных типов хранилищ или моков.
- Service (Infrastructure/background/external system integrations): `internal/service/*`.
  - Фоновые процессы, интеграции с другими внешними системами.
  - выделенные или переиспользуемые логики
  - может использовать другие соседние сервисы `internal/service/*`, а так же доменные сервисы `internal/domain/*/service`
  - не обращается в usecase слой
- Composition: `internal/app/`.
  - Сборка зависимостей, запуск серверов, миграций, фоновых сервисов.

**Правило зависимостей:**
`handler → usecase`.
`usecase → domain service`.
`usecase → service`.
`service → service`.
`service → domain service`.
`domain service → repo`.
Обратные зависимости запрещены.
К repo слою доступ только с `domain service`.

## Хранилища
- Postgres: все доменные entities (см. `migrations/*`).

## API
- gRPC сервисы: `api/proto/airo_v1/*`.
- HTTP-gateway: через grpc-gateway + swagger (`docs/api.swagger.json`).
- DTO маппинг: `internal/handler/grpc/dto`.

## Ошибки и валидация
- Семантические ошибки — через `internal/errs` (см. gRPC interceptor в `internal/app/grpc.go`).
- Валидация параметров — в usecase.
- Нельзя пробрасывать ошибки наружу без wrapping (оборачивать в `fmt.Errorf("...: %w")`).

## Правила изменения кода
- Новые доменные сущности должны иметь `model/`, `service/`, `repo/` и usecase + handler слой. Иногда могут иметь только `model/` если не храним записи.
- gRPC DTO не должны протекать в доменные сервисы.
- `pkg/proto` и `docs/api.swagger.json` — генерируемые файлы (обновляются через Makefile).
