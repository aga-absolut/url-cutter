package workers

import (
	"context"
	"sync"

	"github.com/aga-absolut/url-cutter/internal/repository"
	"go.uber.org/zap"
)

// Worker структура.
type Worker struct {
	deleteChan chan string
	storage    repository.Storage
	logger     *zap.SugaredLogger
	wg         sync.WaitGroup
	size       int
}

// NewWorkerPool создает новый WorkerPool.
func NewWorkerPool(ctx context.Context, deletedChan chan string, storage repository.Storage, size int, logger *zap.SugaredLogger) *Worker {
	w := &Worker{
		deleteChan: deletedChan,
		storage:    storage,
		size:       size,
		logger:     logger,
	}

	w.Start(ctx, size)
	return w
}

// Start создает воркеров.
func (w *Worker) Start(ctx context.Context, size int) {
	w.wg.Add(size)
	for i := 0; i < size; i++ {
		go w.worker(ctx)
	}
}

// worker ждет сигнал контекста или данные из канала.
func (w *Worker) worker(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case shortURL, ok := <-w.deleteChan:
			if !ok {
				return
			}
			if err := w.storage.DeletedFlag(ctx, shortURL); err != nil {
				w.logger.Errorw("Failed to delete %s: %v", shortURL, err)
				return
			}
		}
	}
}

// Stop закрывает канал.
func (w *Worker) Stop() {
	close(w.deleteChan)
	w.wg.Wait()
}
