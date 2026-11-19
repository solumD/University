package internalapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"gitlab.ozon.ru/university/api-2/internal/delivery/http/response"
	"gitlab.ozon.ru/university/api-2/internal/usecase"
	"go.uber.org/zap"
)

func Register(r chi.Router, service *usecase.Service, log *zap.Logger) {
	r.Get("/stats", func(w http.ResponseWriter, req *http.Request) {
		stats, err := service.Stats(req.Context())
		if err != nil {
			response.FromError(w, err, log)

			return
		}

		response.JSON(w, http.StatusOK, stats, log)
	})
}
