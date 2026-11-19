package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gitlab.ozon.ru/university/api-2/config"
	"gitlab.ozon.ru/university/api-2/docs"
	"gitlab.ozon.ru/university/api-2/internal/delivery/http/internalapi"
	"gitlab.ozon.ru/university/api-2/internal/delivery/http/middleware"
	"gitlab.ozon.ru/university/api-2/internal/delivery/http/response"
	v1 "gitlab.ozon.ru/university/api-2/internal/delivery/http/v1"
	v2 "gitlab.ozon.ru/university/api-2/internal/delivery/http/v2"
	"gitlab.ozon.ru/university/api-2/internal/usecase"
	"go.uber.org/zap"
)

func NewRouter(cfg config.Config, service *usecase.Service, log *zap.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logging(log))
	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		response.Fail(w, http.StatusNotFound, "not_found", "route not found", log)
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		response.Fail(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", log)
	})

	r.Get("/docs", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFileFS(w, req, docs.Files, "index.html")
	})

	r.Get("/openapi.yaml", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		http.ServeFileFS(w, req, docs.Files, "openapi.yaml")
	})

	security := middleware.NewSecurity(cfg.APIKey, cfg.RateLimit, log)
	r.Route("/api", func(api chi.Router) {
		api.Use(security.RateLimit, security.Authenticate, middleware.Idempotency(log))
		api.Route("/v1", func(router chi.Router) {
			v1.Register(router, service, log)
		})
		api.Route("/v2", func(router chi.Router) {
			v2.Register(router, service, log)
		})
	})

	if cfg.InternalAPIKey != "" {
		internalSecurity := middleware.NewSecurity(cfg.InternalAPIKey, cfg.RateLimit, log)

		r.Route("/internal/v1", func(api chi.Router) {
			api.Use(internalSecurity.RateLimit, internalSecurity.Authenticate)
			internalapi.Register(api, service, log)
		})
	}

	return r
}
