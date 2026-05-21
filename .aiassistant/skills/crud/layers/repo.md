# Layer: Repository — `internal/domain/<entity>/repo/db/`

Доступ к PostgreSQL через mobone/squirrel/pgx.

## Файлы

| Файл | Назначение |
|---|---|
| `repo.go` | `Repo` struct, `New()`, CRUD-методы |
| `custom.go` | `getConditions`, `allowedSortFields`, кастомные запросы |
| `model/select.go` | модель чтения (value-типы) + `EncodeSelect` |
| `model/upsert.go` | модель записи (pointer-типы) + `DecodeUpsert` |

> DTOs (`EncodeSelect`, `DecodeUpsert`) живут в своих файлах — отдельного `dto.go` нет.

---

## model/select.go — модель чтения + DTO

- Поля **value-типы** (не указатели)
- Nullable колонки в БД → `sql.NullString` / `sql.NullTime`
- `EncodeSelect` — конвертер Select → domain.Main (сигнатура совместима с `lo.Map`)

```go
package model

import (
    "time"

    domainModel "github.com/<module>/internal/domain/<entity>/model"
)

type Select struct {
    Id         string
    CreatedAt  time.Time
    ModifiedAt time.Time
    Active     bool
    Name       string
}

func (m *Select) ListColumnMap() map[string]any {
    return map[string]any{
        "id":          &m.Id,
        "created_at":  &m.CreatedAt,
        "modified_at": &m.ModifiedAt,
        "active":      &m.Active,
        "name":        &m.Name,
    }
}

func (m *Select) PKColumnMap() map[string]any {
    return map[string]any{"id": m.Id}
}

func (m *Select) DefaultSortColumns() []string {
    return []string{"created_at DESC"}
}

// DTO

func EncodeSelect(v *Select, _ int) *domainModel.Main {
    return &domainModel.Main{
        Id:         v.Id,
        CreatedAt:  v.CreatedAt,
        ModifiedAt: v.ModifiedAt,
        Active:     v.Active,
        Name:       v.Name,
    }
}
```

---

## model/upsert.go — модель записи + DTO

- `PKId string` — публичный, не указатель; устанавливается в `repo.go` перед Update/Delete
- `CreateColumnMap` — **не включает** `id` (только бизнес-поля)
- `UpdateColumnMap` — делегирует в `CreateColumnMap`
- `DecodeUpsert` — конвертер domain.Edit → Upsert

```go
package model

import (
    "time"

    domainModel "github.com/<module>/internal/domain/<entity>/model"
)

type Upsert struct {
    PKId string

    ModifiedAt *time.Time
    Active     *bool
    Name       *string
}

func (m *Upsert) CreateColumnMap() map[string]any {
    result := make(map[string]any{}, 10)
    if m.ModifiedAt != nil { result["modified_at"] = *m.ModifiedAt }
    if m.Active != nil     { result["active"] = *m.Active }
    if m.Name != nil       { result["name"] = *m.Name }
    return result
}

func (m *Upsert) UpdateColumnMap() map[string]any {
    return m.CreateColumnMap()
}

func (m *Upsert) PKColumnMap() map[string]any {
    return map[string]any{"id": m.PKId}
}

func (m *Upsert) ReturningColumnMap() map[string]any {
    return map[string]any{}
}

// DTO

func DecodeUpsert(v *domainModel.Edit) *Upsert {
    return &Upsert{
        ModifiedAt: v.ModifiedAt,
        Active:     v.Active,
        Name:       v.Name,
    }
}
```

---

## repo.go — Repo, конструктор, CRUD

```go
package db

import (
    "context"
    "fmt"

    "github.com/<module>/internal/domain/<entity>/model"
   
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/mechta-market/mobone/v2"
    "github.com/samber/lo"

    commonRepoPg "github.com/<module>/internal/domain/common/repo/pg"
    moboneTools "github.com/mechta-market/mobone/v2/tools"
    repoModel "github.com/<module>/internal/domain/<entity>/repo/db/model"
)

type Repo struct {
    *commonRepoPg.Base
    ModelStore *mobone.ModelStore
}

func New(con *pgxpool.Pool) *Repo {
    base := commonRepoPg.NewBase(con)
    return &Repo{
        Base: base,
        ModelStore: &mobone.ModelStore{
            Con:       base.Con,
            QB:        base.QB,
            TableName: "<table_name>",
        },
    }
}

func (r *Repo) List(ctx context.Context, pars *model.ListReq) ([]*model.Main, int64, error) {
    conditions, conditionExps := r.getConditions(pars)
    sort := moboneTools.ConstructSortColumns(allowedSortFields, pars.Sort)
    items := make([]*repoModel.Select, 0)

    totalCount, err := r.ModelStore.List(ctx, mobone.ListParams{
        Conditions:           conditions,
        ConditionExpressions: conditionExps,
        Page:                 pars.Page,
        PageSize:             pars.PageSize,
        WithTotalCount:       pars.WithTotalCount,
        OnlyCount:            pars.OnlyCount,
        Sort:                 sort,
    }, func(add bool) mobone.ListModelI {
        item := &repoModel.Select{}
        if add {
            items = append(items, item)
        }
        return item
    })
    if err != nil {
        return nil, 0, fmt.Errorf("ModelStore.List: %w", err)
    }
    return lo.Map(items, repoModel.EncodeSelect), totalCount, nil
}

func (r *Repo) Get(ctx context.Context, id string) (*model.Main, bool, error) {
    m := &repoModel.Select{Id: id}
    found, err := r.ModelStore.Get(ctx, m)
    if err != nil {
        return nil, false, fmt.Errorf("ModelStore.Get: %w", err)
    }
    if !found {
        return nil, false, nil
    }
    return repoModel.EncodeSelect(m, 0), true, nil
}

func (r *Repo) Create(ctx context.Context, obj *model.Edit) error {
    m := repoModel.DecodeUpsert(obj)
    if err := r.ModelStore.Create(ctx, m); err != nil {
        return fmt.Errorf("ModelStore.Create: %w", err)
    }
    return nil
}

func (r *Repo) Update(ctx context.Context, id string, obj *model.Edit) error {
    m := repoModel.DecodeUpsert(obj)
    m.PKId = id
    if err := r.ModelStore.Update(ctx, m); err != nil {
        return fmt.Errorf("ModelStore.Update: %w", err)
    }
    return nil
}

func (r *Repo) Delete(ctx context.Context, id string) error {
    m := &repoModel.Upsert{PKId: id}
    if err := r.ModelStore.Delete(ctx, m); err != nil {
        return fmt.Errorf("ModelStore.Delete: %w", err)
    }
    return nil
}
```

---

## custom.go — фильтры, сортировка, сложные запросы

```go
package db

import (
    sq "github.com/Masterminds/squirrel"
    "github.com/<module>/internal/domain/<entity>/model"
)

var allowedSortFields = map[string]string{
    "id":          "id",
    "created_at":  "created_at",
    "modified_at": "modified_at",
    "name":        "name",
}

func (r *Repo) getConditions(pars *model.ListReq) (map[string]any, map[string][]any) {
    conditions := make(map[string]any, 10)
    conditionExps := make(map[string][]any, 10)

    if pars == nil {
        return conditions, conditionExps
    }

    // Точное совпадение:
    // if pars.Active != nil { conditions["active"] = *pars.Active }

    // ILIKE поиск:
    // if pars.Name != nil { conditionExps["name ILIKE ?"] = []any{"%" + *pars.Name + "%"} }

    // IN по массиву:
    // if pars.IDs != nil { condition["id"] = pars.IDs }

    return conditions, conditionExps
}

// Bulk операция (squirrel)
func (r *Repo) DeleteByFilter(ctx context.Context, field string) error {
    query, args, err := r.QB.
        Delete("<table_name>").
        Where(sq.Eq{"<column>": field}).
        ToSql()
    if err != nil {
        return fmt.Errorf("DeleteByFilter build query: %w", err)
    }
    if _, err = r.Con.Exec(ctx, query, args...); err != nil {
        return fmt.Errorf("DeleteByFilter: %w", err)
    }
    return nil
}
```

---

## Трассировка (опционально)

Если проект использует OpenTracing, оборачивай каждый публичный метод:

```go
func (r *Repo) MethodName(ctx context.Context, ...) (_ RetType, finalError error) {
    tracingSpan, ctx := opentracing.StartSpanFromContext(ctx, "<entity>.repo.DB.MethodName")
    defer tracingSpan.Finish()
    defer func() {
        if finalError != nil {
            tracingSpan.SetTag("error", true)
            tracingSpan.LogKV("error", finalError.Error())
        }
    }()
    // ...
}
```
