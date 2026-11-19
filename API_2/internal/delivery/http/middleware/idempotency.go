package middleware

import (
	"net/http"
	"strings"

	"gitlab.ozon.ru/university/api-2/internal/delivery/http/response"
	"gitlab.ozon.ru/university/api-2/internal/model"
	"go.uber.org/zap"
)

func Idempotency(log *zap.Logger) func(http.Handler) http.Handler {
	middleware := func(next http.Handler) http.Handler {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				key := r.Header.Get("Idempotency-Key")
				validKey := len(key) >= 1 && len(key) <= 128 && strings.IndexFunc(key, invalidKeyCharacter) < 0

				if len(r.Header.Values("Idempotency-Key")) != 1 || !validKey {
					response.Fail(w, http.StatusBadRequest, "invalid_request", "Idempotency-Key must contain 1 to 128 visible ASCII characters without spaces", log)

					return
				}
			}

			next.ServeHTTP(w, r)
		})

		return handler
	}

	return middleware
}

func CommandKey(r *http.Request) model.CommandKey {
	client, _ := r.Context().Value(clientKey{}).(string)
	key := model.CommandKey{
		Scope: client + " " + r.Method + " " + r.URL.Path,
		Key:   r.Header.Get("Idempotency-Key"),
	}

	return key
}

func invalidKeyCharacter(character rune) bool {
	return character < 33 || character > 126
}
