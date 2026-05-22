package heroname

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yerlan/dota2/internal/service/opendota"
)

type SourceI interface {
	GetHeroStats(ctx context.Context) ([]*opendota.HeroStat, error)
}

type Resolver struct {
	src SourceI
	ttl time.Duration

	mu    sync.Mutex
	names map[int]string
	at    time.Time
}

func New(src SourceI, ttl time.Duration) *Resolver {
	return &Resolver{src: src, ttl: ttl}
}

// Name возвращает локализованное имя героя по hero_id; при недоступном
// справочнике или неизвестном id возвращает строку "hero <N>".
func (r *Resolver) Name(ctx context.Context, heroID int) string {
	r.ensure(ctx)
	r.mu.Lock()
	defer r.mu.Unlock()
	if n, ok := r.names[heroID]; ok {
		return n
	}
	return fmt.Sprintf("hero %d", heroID)
}

func (r *Resolver) ensure(ctx context.Context) {
	r.mu.Lock()
	fresh := r.names != nil && time.Since(r.at) < r.ttl
	r.mu.Unlock()
	if fresh {
		return
	}
	stats, err := r.src.GetHeroStats(ctx)
	if err != nil || stats == nil {
		return
	}
	m := make(map[int]string, len(stats))
	for _, h := range stats {
		m[h.ID] = h.LocalizedName
	}
	r.mu.Lock()
	r.names = m
	r.at = time.Now()
	r.mu.Unlock()
}
