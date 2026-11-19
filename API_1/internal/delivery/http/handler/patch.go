package handler

import (
	"bytes"
	"encoding/json"
	"net/http"

	"gitlab.ozon.ru/university/api-1/internal/model"
)

func decodePatch(w http.ResponseWriter, r *http.Request, destination any) error {
	var data json.RawMessage

	err := decode(w, r, &data)
	if err != nil {
		return err
	}

	var fields map[string]json.RawMessage

	err = json.Unmarshal(data, &fields)
	if err != nil || len(fields) == 0 {
		validationErr := model.Invalid("PATCH body must be a nonempty JSON object")

		return validationErr
	}

	for _, value := range fields {
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			validationErr := model.Invalid("PATCH fields must not be null")

			return validationErr
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	err = decoder.Decode(destination)
	if err != nil {
		validationErr := model.Invalid("PATCH body must contain only valid fields")

		return validationErr
	}

	return nil
}
