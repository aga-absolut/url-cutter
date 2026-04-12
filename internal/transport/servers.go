package transport

import (
	"net"
	"net/http"

	"github.com/aga-absolut/url-cutter/internal/cert"
	"github.com/aga-absolut/url-cutter/internal/config"
	"github.com/aga-absolut/url-cutter/internal/service"
	grpcserver "github.com/aga-absolut/url-cutter/internal/transport/grpc/grpc_server"
	pb "github.com/aga-absolut/url-cutter/internal/transport/grpc/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func StartHTTPServer(cfg *config.Config, handler http.Handler, logger *zap.SugaredLogger) *http.Server {
	server := &http.Server{
		Addr:    cfg.HTTPServerAddress,
		Handler: handler,
	}

	go func() {
		logger.Infow("Starting HTTP server", "addr", cfg.HTTPServerAddress)

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

	return server
}

// grpcurl -plaintext -d '{\"original_url\": \"https://google.com\"}' localhost:3200 urlcutter.URLCutter.PostHandler
// grpcurl -plaintext -d '{\"short_url\": \"940689ec\"}' localhost:3200 urlcutter.URLCutter.GetHandler
func StartGRPCServer(cfg *config.Config, service service.Service, logger *zap.SugaredLogger) *grpc.Server {
	server := grpc.NewServer()
	pb.RegisterURLCutterServer(server, grpcserver.NewURLCutterServer(service))
	reflection.Register(server)

	go func() {
		logger.Infow("Starting gRPC server", "addr", cfg.GRPCServerAddress)

		listen, err := net.Listen("tcp", cfg.GRPCServerAddress)
		if err != nil {
			logger.Fatalw("failed to listen on gRPC port", "port", cfg.GRPCServerAddress, "error", err)
		}

		if err := server.Serve(listen); err != nil {
			logger.Errorw("gRPC server stopped", "error", err)
		}
	}()

	return server
}
