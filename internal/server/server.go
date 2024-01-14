package server

import (
	"context"
	"net/http"

	"github.com/adminsemy/WebSocket/internal/server/handler"
	"go.uber.org/zap"
)

func Run(ctx context.Context, logger *zap.Logger) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	webSocketUpgrader := handler.NewManager(ctx)
	go webSocketUpgrader.RemoveClient()

	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.HandleFunc("/login", webSocketUpgrader.Login)
	http.HandleFunc("/ws", webSocketUpgrader.ServeWS)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
