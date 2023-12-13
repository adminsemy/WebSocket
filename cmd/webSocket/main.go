package main

import (
	"context"

	"github.com/adminsemy/WebSocket/internal/server"
	"go.uber.org/zap"
)

func main() {
	// 1. Initialize the logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	server.Run(context.Background(), logger)
}
