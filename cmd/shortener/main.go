package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aga-absolut/url-cutter/internal/cert"
	"github.com/aga-absolut/url-cutter/internal/config"
	grpc_server "github.com/aga-absolut/url-cutter/internal/grpc"
	"github.com/aga-absolut/url-cutter/internal/handler"
	"github.com/aga-absolut/url-cutter/internal/router"
	"github.com/aga-absolut/url-cutter/internal/service"
	"github.com/aga-absolut/url-cutter/internal/storage"
	"github.com/aga-absolut/url-cutter/internal/storage/postgreSQL/database"
	"github.com/aga-absolut/url-cutter/internal/workers"
	"github.com/aga-absolut/url-cutter/middleware/logger"
	pb "github.com/aga-absolut/url-cutter/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	HTTPserver := &http.Server{
		Addr:    cfg.HTTPServerAddress,
		Handler: router,
	}

	go func() {
		logger.Infow("Starting HTTP server", "addr", cfg.HTTPServerAddress)

		if cfg.EnableHTTPS {
			certFile := "server.crt"
			keyFile := "server.key"

			if err := cert.GenerateCertificate(certFile, keyFile); err != nil {
				logger.Fatalw("failed generate certificate", "error", err)
			}

			if err := HTTPserver.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				logger.Fatalw("create HTTPS server error", "error", err)
			}
		} else {
			if err := HTTPserver.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatalw("create HTTP server error", "error", err)
			}
		}
	}()

	// grpcurl -plaintext -d '{\"original_url\": \"https://google.com\"}' localhost:3200 urlcutter.URLCutter.PostHandler
	// grpcurl -plaintext -d '{\"short_url\": \"940689ec\"}' localhost:3200 urlcutter.URLCutter.GetHandler

	gRPCserver := grpc.NewServer()
	pb.RegisterURLCutterServer(gRPCserver, grpc_server.NewURLCutterServer(service))
	reflection.Register(gRPCserver)

	go func() {
		logger.Infow("Starting gRPC server", "addr", cfg.GRPCServerAddress)

		listen, err := net.Listen("tcp", cfg.GRPCServerAddress)
		if err != nil {
			logger.Fatalw("failed to listen on gRPC port :3200", "error", err)
		}

		if err := gRPCserver.Serve(listen); err != nil {
			logger.Errorw("gRPC server stopped", "error", err)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutdown signal received")

	ShutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := HTTPserver.Shutdown(ShutdownCtx); err != nil {
		logger.Errorw("Server shutdown error", "Error", err)
	}

	gRPCserver.GracefulStop()
	worker.Stop()
	logger.Info("Application stopped successfully")
}
