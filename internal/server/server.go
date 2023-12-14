package server

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

func Run(ctx context.Context, logger *zap.Logger) {
	http.Handle("/", http.FileServer(http.Dir("./web")))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
