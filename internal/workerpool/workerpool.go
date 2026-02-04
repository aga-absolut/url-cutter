package workerpool

import (
	"context"
	"log"
	"sync"

	"github.com/aga-absolut/url-cutter/internal/repository"
)

type Worker struct {
	deleteChan chan string
	ctx        context.Context
	storage    repository.Storage
	wg         sync.WaitGroup
	size       int
}

func NewWorkerPool(ctx context.Context, deletedChan chan string, storage repository.Storage, size int) *Worker{
	w := &Worker{
		ctx:        ctx,
		deleteChan: deletedChan,
		storage: storage,
		size: size,
	}

	w.wg.Add(size)
	for i := 0; i < size; i++ {
		go w.worker()
	}
	return w
}

func (w *Worker) worker() {
	defer w.wg.Done()
	for {
		select {
		case <-w.ctx.Done():
			return
		case shortURL, ok := <-w.deleteChan:
			if !ok {
				return
			}
			if err := w.storage.DeletedFlag(w.ctx, shortURL); err != nil {
				log.Printf("Failed to delete %s: %v", shortURL, err)
			}
		}
	}
}

func (w *Worker) Stop() {
	close(w.deleteChan)
	w.wg.Wait()
}
