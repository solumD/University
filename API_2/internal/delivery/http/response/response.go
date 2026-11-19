package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"gitlab.ozon.ru/university/api-2/internal/model"
	"go.uber.org/zap"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, value any, log *zap.Logger) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Error("encode response", zap.Error(err))
		status = http.StatusInternalServerError
		data = []byte(`{"code":"internal_error","message":"internal server error"}`)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err = w.Write(append(data, '\n')); err != nil {
		log.Warn("write response", zap.Error(err))
	}
}

func Fail(w http.ResponseWriter, status int, code, message string, log *zap.Logger) {
	JSON(w, status, Error{Code: code, Message: message}, log)
}

func FromError(w http.ResponseWriter, err error, log *zap.Logger) {
	if domain, ok := errors.AsType[*model.Error](err); ok {
		status := http.StatusBadRequest

		switch domain.Code {
		case "project_not_found", "task_not_found":
			status = http.StatusNotFound
		case "project_not_empty", "idempotency_conflict":
			status = http.StatusConflict
		}

		Fail(w, status, domain.Code, domain.Message, log)

		return
	}

	log.Error("request failed", zap.Error(err))
	Fail(w, http.StatusInternalServerError, "internal_error", "internal server error", log)
}
