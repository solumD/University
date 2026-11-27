package library

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

var ErrLoginRequired = errors.New("сначала войдите в меню читателя")

type APIError struct {
	Status     int
	Detail     string
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	var message string

	switch e.Status {
	case http.StatusUnauthorized:
		message = "проверьте учётные данные и войдите снова"
	case http.StatusNotFound:
		message = "ресурс не найден"
	case http.StatusUnprocessableEntity:
		message = "ошибка проверки данных"
	case http.StatusTooManyRequests:
		message = "лимит запросов исчерпан"
		if e.RetryAfter > 0 {
			message += fmt.Sprintf(", повторите через %.0f с", math.Ceil(e.RetryAfter.Seconds()))
		}
	default:
		message = http.StatusText(e.Status)
	}

	if e.Detail != "" {
		message += " (" + e.Detail + ")"
	}

	text := fmt.Sprintf("HTTP %d — %s", e.Status, message)

	return text
}

func readError(response *resty.Response) *APIError {
	failure := &APIError{Status: response.StatusCode()}

	if raw := response.Header().Get("Retry-After"); raw != "" {
		delay, err := time.ParseDuration(raw + "s")
		if err == nil && delay > 0 {
			failure.RetryAfter = delay
		}
	}

	var body struct {
		Detail json.RawMessage `json:"detail"`
		Error  string          `json:"error"`
	}

	if err := json.Unmarshal(response.Body(), &body); err != nil {
		return failure
	}

	failure.Detail = body.Error
	if len(body.Detail) == 0 {
		return failure
	}

	var message string
	if err := json.Unmarshal(body.Detail, &message); err == nil {
		failure.Detail = message

		return failure
	}

	var issues []struct {
		Location []any  `json:"loc"`
		Message  string `json:"msg"`
	}

	if err := json.Unmarshal(body.Detail, &issues); err != nil {
		return failure
	}

	messages := make([]string, 0, len(issues))
	for _, issue := range issues {
		field := make([]string, 0, len(issue.Location))
		for _, part := range issue.Location {
			field = append(field, fmt.Sprint(part))
		}

		messages = append(messages, strings.Join(field, ".")+" — "+issue.Message)
	}

	failure.Detail = strings.Join(messages, ", ")

	return failure
}
