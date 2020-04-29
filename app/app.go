package app

import (
	"context"
	files_api "github.com/go-project/proto/generated"
	"github.com/go-project/services"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type App struct {
	err    chan error
	logger *zap.Logger
}

func New(ctx context.Context, address string) (*App, error) {

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	errChan := make(chan error)

	app := &App{
		logger: logger,
		err:    errChan,
	}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()

	files_api.RegisterFilesSvcServer(grpcServer, services.NewFilesSvc(logger))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server fatal error", zap.Error(err))
			errChan <- err
		}
	}()

	go func() {
		appCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		defer func() {
			defer logger.Info("application is stopped")
			defer close(errChan)
			defer grpcServer.Stop()

			logger.Info("stopping application...")
		}()

		logger.Info("application is started")
		select {
		case <-appCtx.Done():
		}
	}()

	return app, nil
}

func (a *App) GetLogger() *zap.Logger {
	return a.logger
}

func (a *App) GetErr() chan error {
	return a.err
}
