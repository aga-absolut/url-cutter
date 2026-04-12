package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/service"
	"github.com/aga-absolut/url-cutter/internal/storage"
	"github.com/aga-absolut/url-cutter/internal/storage/postgreSQL/database"
	"github.com/aga-absolut/url-cutter/internal/transport"
	"github.com/aga-absolut/url-cutter/internal/transport/http/handler"
	"github.com/aga-absolut/url-cutter/internal/transport/http/middleware/logger"
	"github.com/aga-absolut/url-cutter/internal/transport/http/router"
	"github.com/aga-absolut/url-cutter/internal/workers"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	deleteChan := make(chan string, 10)
	cfg := config.NewConfig()
	logger := logger.NewLogger()
	if cfg.DBDSN != "" {
		if err := database.InitMigrations(cfg, logger); err != nil {
			logger.Fatalw("don`t create migrations", "error", err)
		}
	}
	storage := storage.NewStorage(cfg, logger)
	worker := workers.NewWorkerPool(ctx, deleteChan, storage, config.SizeWorkers, logger)
	service := service.NewService(cfg, storage, deleteChan)
	handler := handler.NewHandler(service, logger)
	router := router.NewRouter(handler)

	// Запуск серверов
	httpServer := transport.StartHTTPServer(cfg, router, logger)
	grpcServer := transport.StartGRPCServer(cfg, service, logger)

	<-ctx.Done()
	logger.Info("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Errorw("Server shutdown error", "Error", err)
	}

	grpcServer.GracefulStop()
	worker.Stop()
	logger.Info("Application stopped successfully")
}
