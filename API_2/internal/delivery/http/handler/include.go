package handler

import (
	"net/http"
	"net/url"
	"slices"
	"strings"

	"gitlab.ozon.ru/university/api-2/internal/model"
)

var (
	projectFields = []string{"id", "name"}
	taskFields    = []string{"id", "project_id", "title", "completed", "description"}
)

func parseInclude(r *http.Request, allowed []string) ([]string, error) {
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		validationErr := model.Invalid("invalid query string")

		return nil, validationErr
	}

	values, exists := query["include"]
	if !exists {
		return nil, nil
	}

	if len(values) != 1 {
		validationErr := model.Invalid("include must not be repeated")

		return nil, validationErr
	}

	names := strings.Split(values[0], ",")
	fields := make([]string, 0, len(names))

	for _, name := range names {
		name = strings.TrimSpace(name)
		if !slices.Contains(allowed, name) || slices.Contains(fields, name) {
			validationErr := model.Invalid("include must contain known, nonempty and unique field names")

			return nil, validationErr
		}

		fields = append(fields, name)
	}

	return fields, nil
}
