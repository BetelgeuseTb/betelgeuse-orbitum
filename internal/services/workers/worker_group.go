package workers

import (
	"context"
	"sync"
)

type Worker interface {
	Start(ctx context.Context) error
}

type WorkerGroup struct {
	workers []Worker
	wg      sync.WaitGroup
	cancel  context.CancelFunc
}

func NewWorkerGroup() *WorkerGroup {
	return &WorkerGroup{
		workers: make([]Worker, 0),
	}
}

func (g *WorkerGroup) Add(worker Worker) {
	g.workers = append(g.workers, worker)
}

func (g *WorkerGroup) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	g.cancel = cancel

	for _, worker := range g.workers {
		g.wg.Add(1)
		go func(w Worker) {
			defer g.wg.Done()
			_ = w.Start(ctx)
		}(worker)
	}
}

func (g *WorkerGroup) Stop(ctx context.Context) error {
	if g.cancel != nil {
		g.cancel()
	}

	done := make(chan struct{})
	go func() {
		g.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
