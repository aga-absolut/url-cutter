package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aga-absolut/url-cutter/internal/cert"
	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/handler"
	"github.com/aga-absolut/url-cutter/internal/router"
	"github.com/aga-absolut/url-cutter/internal/storage"
	"github.com/aga-absolut/url-cutter/internal/storage/postgreSQL/database"
	"github.com/aga-absolut/url-cutter/internal/workers"
	"github.com/aga-absolut/url-cutter/middleware/logger"
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
	handler := handler.NewHandler(cfg, storage, logger, deleteChan)
	router := router.NewRouter(handler)

	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}

	go func() {
		logger.Infow("Starting server", "addr", cfg.ServerAddress)

		if cfg.EnableHTTPS {
			certFile := "server.crt"
			keyFile := "server.key"

			if err := cert.GenerateCertificate(certFile, keyFile); err != nil {
				logger.Fatalw("failed generate certificate", "error", err)
			}

			if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				logger.Fatalw("create HTTPS server error", "error", err)
			}
		} else {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatalw("create HTTP server error", "error", err)
			}
		}
	}()

	<-ctx.Done()
	logger.Info("Shutdown signal received")

	ShutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ShutdownCtx); err != nil {
		logger.Errorw("Server shutdown error", "Error", err)
	}

	worker.Stop()
	logger.Info("Application stopped successfully")
}
