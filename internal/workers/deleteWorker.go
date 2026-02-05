package workers

import (
	"context"
	"sync"

	"github.com/aga-absolut/url-cutter/internal/storage/postgreSQL/database"
	"go.uber.org/zap"
)

type Worker struct {
	deleteChan chan string
	SQLDB      *database.DBPostgreSQL
	logger     zap.SugaredLogger
	wg         sync.WaitGroup
	size       int
}

func NewWorkerPool(ctx context.Context, deletedChan chan string, SQLDB *database.DBPostgreSQL, size int, logger zap.SugaredLogger) *Worker {
	w := &Worker{
		deleteChan: deletedChan,
		SQLDB:      SQLDB,
		size:       size,
		logger:     logger,
	}

	w.wg.Add(size)
	for i := 0; i < size; i++ {
		go w.worker(ctx)
	}
	return w
}

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
			if err := w.SQLDB.DeletedFlag(ctx, shortURL); err != nil {
				w.logger.Errorw("Failed to delete %s: %v", shortURL, err)
				return
			}
		}
	}
}

func (w *Worker) Stop() {
	close(w.deleteChan)
	w.wg.Wait()
}
