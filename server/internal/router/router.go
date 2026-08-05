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
		mux.HandleFunc("POST /api/v1/auth/forgot-password", authHandler.ForgotPassword)
		mux.HandleFunc("POST /api/v1/auth/reset-password", authHandler.ResetPassword)
		mux.Handle("GET /api/v1/auth/me", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.Me)))
		mux.Handle("PATCH /api/v1/auth/me/avatar", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.UpdateAvatar)))
		mux.Handle("DELETE /api/v1/auth/me/avatar", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.ResetAvatar)))
		mux.Handle("POST /api/v1/auth/change-password", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.ChangePassword)))
		mux.Handle("POST /api/v1/auth/upgrade-to-developer", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.UpgradeToDeveloper)))
		mux.Handle("GET /api/v1/auth/developer-application", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.MyDeveloperApplication)))
		mux.Handle("GET /api/v1/admin/developer-applications", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.ListDeveloperApplications)))
		mux.Handle("POST /api/v1/admin/developer-applications/{id}/approve", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.ApproveDeveloperApplication)))
		mux.Handle("POST /api/v1/admin/developer-applications/{id}/reject", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.RejectDeveloperApplication)))
		mux.Handle("POST /api/v1/apps", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.CreateApp)))
		mux.Handle("GET /api/v1/developer/apps", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.ListDeveloperApps)))
		mux.Handle("GET /api/v1/developer/apps/{id}", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.GetDeveloperApp)))
		mux.Handle("PUT /api/v1/developer/apps/{id}", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.UpdateDeveloperApp)))
		mux.Handle("POST /api/v1/developer/apps/{id}/delist", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.DelistDeveloperApp)))
		mux.Handle("GET /api/v1/admin/apps", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.ListAdminApps)))
		mux.Handle("POST /api/v1/admin/apps/{id}/approve", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.ApproveApp)))
		mux.Handle("POST /api/v1/admin/apps/{id}/reject", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.RejectApp)))
		mux.HandleFunc("GET /api/v1/marketplace/apps", authHandler.ListMarketplaceApps)
		mux.HandleFunc("GET /api/v1/marketplace/apps/{slug}", authHandler.GetMarketplaceApp)
		mux.Handle("GET /api/v1/marketplace/apps/{slug}/favorite", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.GetMarketplaceFavorite)))
		mux.Handle("POST /api/v1/marketplace/apps/{slug}/favorite", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.ToggleMarketplaceFavorite)))
		mux.Handle("POST /api/v1/marketplace/apps/{slug}/checkout", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.Checkout)))
		mux.Handle("POST /api/v1/orders/{id}/confirm-payment", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.ConfirmPayment)))
		mux.Handle("GET /api/v1/me/orders", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.ListMyOrders)))
		mux.Handle("GET /api/v1/me/entitlements", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.ListMyEntitlements)))
		mux.Handle("GET /api/v1/entitlements/{id}/delivery", middleware.JWT(configured.auth.JWTSecret)(http.HandlerFunc(authHandler.GetDelivery)))
		mux.Handle("GET /api/v1/uploads/avatars/", http.StripPrefix("/api/v1/uploads/avatars/", http.FileServer(http.Dir(configured.auth.AvatarUploadDir))))
	}

	return middleware.Logging(logger)(middleware.Recovery(logger)(mux))
}
