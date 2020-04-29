package main

import (
	"context"
	"github.com/go-project/app"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"time"
)

const address = ":9876"

func main() {

	defer time.Sleep(1500 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := app.New(ctx, address)
	if err != nil {
		panic(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	logger := app.GetLogger()

	select {
	case <-stop:
		logger.Info("application was interrupted")
	case err := <-app.GetErr():
		logger.Panic("fatal error occurred", zap.Error(err))
	}
}
