package router

import (
	"log/slog"
	"net/http"

	"devkit/server/internal/handler"
	"devkit/server/internal/middleware"
)

func New(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", handler.Health)

	return middleware.Logging(logger)(middleware.Recovery(logger)(mux))
}
