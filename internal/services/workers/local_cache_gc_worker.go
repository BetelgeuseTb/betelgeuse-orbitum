package workers

import (
	"context"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/services/cache/local"
	"log/slog"
	"time"
)

type GCWorker struct {
	manager  *local.Manager
	interval time.Duration
	logger   *slog.Logger
}

type GCWorkerOption func(*GCWorker)

func WithLogger(logger *slog.Logger) GCWorkerOption {
	return func(w *GCWorker) {
		w.logger = logger
	}
}

func WithInterval(interval time.Duration) GCWorkerOption {
	return func(w *GCWorker) {
		w.interval = interval
	}
}

func NewGCWorker(manager *local.Manager, opts ...GCWorkerOption) *GCWorker {
	w := &GCWorker{
		manager:  manager,
		interval: 1 * time.Minute,
		logger:   slog.Default(),
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

func (w *GCWorker) Start(ctx context.Context) error {
	if w.interval <= 0 {
		return nil
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Info("garbage collector started",
		slog.Duration("interval", w.interval))

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("garbage collector stopped")
			return ctx.Err()
		case <-ticker.C:
			w.collect()
		}
	}
}

func (w *GCWorker) collect() {
	start := time.Now()
	totalEvicted := 0
	totalSize := 0

	caches := w.manager.GetAllCaches()

	for _, cache := range caches {
		evicted := cache.EvictExpired()
		size := cache.Size()

		totalEvicted += evicted
		totalSize += size

		if evicted > 0 {
			w.logger.Debug("local cache garbage collected",
				slog.String("cache", cache.GetCacheName()),
				slog.Int("evicted", evicted),
				slog.Int("remaining", size))
		}
	}

	if totalEvicted > 0 {
		w.logger.Info("garbage collection completed",
			slog.Int("total_evicted", totalEvicted),
			slog.Int("total_items", totalSize),
			slog.Duration("duration", time.Since(start)))
	}
}

func (w *GCWorker) CollectNow() {
	w.collect()
}
