package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gitlab.ozon.ru/university/api-1/internal/delivery/http/response"
	"gitlab.ozon.ru/university/api-1/internal/model"
	"gitlab.ozon.ru/university/api-1/internal/usecase"
	"go.uber.org/zap"
)

type Handler struct {
	service   *usecase.Service
	log       *zap.Logger
	paginated bool
}

func New(service *usecase.Service, log *zap.Logger, paginated bool) *Handler {
	handler := &Handler{service: service, log: log, paginated: paginated}

	return handler
}

func decode(w http.ResponseWriter, r *http.Request, dst any) error {
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		validationErr := model.Invalid("Content-Type must be application/json")

		return validationErr
	}

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		validationErr := model.Invalid("body must be a JSON object with valid fields, at most 64 KiB")

		return validationErr
	}

	var extra any

	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		validationErr := model.Invalid("body must contain exactly one JSON object")

		return validationErr
	}

	return nil
}

func identifier(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		validationErr := model.Invalid("id must be a positive integer")

		return 0, validationErr
	}

	return id, nil
}

func (h *Handler) fail(w http.ResponseWriter, err error) {
	response.FromError(w, err, h.log)
}

func (h *Handler) created(w http.ResponseWriter, r *http.Request, id int64, value any) {
	w.Header().Set("Location", r.URL.Path+"/"+strconv.FormatInt(id, 10))
	response.JSON(w, http.StatusCreated, value, h.log)
}

func (h *Handler) filter(r *http.Request, tasks bool) (model.Filter, error) {
	filter := model.Filter{Limit: -1}
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		validationErr := model.Invalid("invalid query string")

		return filter, validationErr
	}

	if h.paginated {
		filter.Limit = 20
	}

	for name, values := range query {
		if len(values) != 1 {
			validationErr := model.Invalid("query parameters must not be repeated")

			return filter, validationErr
		}

		value := values[0]

		switch {
		case name == "limit" && h.paginated:
			filter.Limit, err = strconv.Atoi(value)
			if err != nil || filter.Limit < 1 || filter.Limit > 100 {
				validationErr := model.Invalid("limit must be between 1 and 100")

				return filter, validationErr
			}
		case name == "offset" && h.paginated:
			filter.Offset, err = strconv.Atoi(value)
			if err != nil || filter.Offset < 0 {
				validationErr := model.Invalid("offset must be nonnegative")

				return filter, validationErr
			}
		case name == "project_id" && tasks:
			id, parseErr := strconv.ParseInt(value, 10, 64)
			if parseErr != nil || id < 1 {
				validationErr := model.Invalid("project_id must be positive")

				return filter, validationErr
			}

			filter.ProjectID = new(id)
		case name == "completed" && tasks:
			if value != "true" && value != "false" {
				validationErr := model.Invalid("completed must be true or false")

				return filter, validationErr
			}

			filter.Completed = new(value == "true")
		default:
			validationErr := model.Invalid("unsupported query parameter")

			return filter, validationErr
		}
	}

	return filter, nil
}
