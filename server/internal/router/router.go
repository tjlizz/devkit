package router

import (
	"database/sql"
	"log/slog"
	"net/http"

	"devkit/server/internal/config"
	"devkit/server/internal/handler"
	"devkit/server/internal/middleware"
)

type Option func(*options)

type options struct {
	db   *sql.DB
	auth config.AuthConfig
}

func WithAuth(db *sql.DB, cfg config.AuthConfig) Option {
	return func(options *options) {
		options.db = db
		options.auth = cfg
	}
}

func New(logger *slog.Logger, routerOptions ...Option) http.Handler {
	var configured options
	for _, option := range routerOptions {
		option(&configured)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", handler.Health)
	if configured.db != nil {
		authHandler := handler.NewAuth(configured.db, configured.auth, logger)
		mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
		mux.HandleFunc("GET /api/v1/auth/activate", authHandler.Activate)
		mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	}

	return middleware.Logging(logger)(middleware.Recovery(logger)(mux))
}
