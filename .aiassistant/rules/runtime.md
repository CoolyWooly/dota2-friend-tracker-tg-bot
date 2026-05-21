---
apply: always
---

# Runtime и конфигурация

## Запуск
- Entry: `cmd/main.go` → `internal/app.App`.
- На старте выполняются:
  - загрузка env (autoload .env),
  - настройка логгера/метрик/трейсинга,
  - pgx pool,
  - миграции (`internal/app/migration.go`),
  - запуск gRPC + HTTP-gateway,

## Переменные окружения
Описаны в `internal/config/config.go`

Примеры: `.env.example`, `.migrate_scripts.example`.

## Метрики и трассировка
- Prometheus метрики доступны на `/metrics` при `WITH_METRICS=true`.
- Трейсинг Jaeger включается при `WITH_TRACING=true` и `JAEGER_ADDRESS`.

## Документация и healthcheck
- `/healthcheck` — HTTP healthcheck (200 OK).
- `/docs/*` — статические docs + swagger (`docs/api.swagger.json`).

## Сборка
- `make build` создаёт бинарник `cmd/build/svc`.
- Dockerfile копирует бинарник, `docs/` и `migrations/` в `/app`.
