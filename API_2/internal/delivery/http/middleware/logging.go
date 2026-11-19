package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func Logging(log *zap.Logger) func(http.Handler) http.Handler {
	middleware := func(next http.Handler) http.Handler {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			writer := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(writer, r)

			log.Info("http request", zap.String("method", r.Method),
				zap.String("route", chi.RouteContext(r.Context()).RoutePattern()),
				zap.Int("status", writer.Status()), zap.Duration("duration", time.Since(start)))
		})

		return handler
	}

	return middleware
}
