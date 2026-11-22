package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/google/uuid"
	"gitlab.ozon.ru/university/broker-4/internal/model"
)

const MaxRequestSize = 64 << 10

type Request struct {
	model.Command
	Auth string `json:"auth"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	CorrelationID string `json:"correlation_id"`
	Status        string `json:"status"`
	Data          any    `json:"data"`
	Error         *Error `json:"error"`
}

func Decode(body []byte) (Request, error) {
	var request Request
	if len(body) > MaxRequestSize {
		err := model.Protocol("invalid_message", "message exceeds 64 KiB")
		return request, err
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		invalid := model.Protocol("invalid_message", "message must contain valid request fields")
		return request, invalid
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		invalid := model.Protocol("invalid_message", "message must contain one JSON object")
		return request, invalid
	}

	id, err := uuid.Parse(request.ID)
	if err != nil || id.String() != request.ID || id == uuid.Nil {
		invalid := model.Protocol("invalid_message", "id must be a nonzero canonical UUID")
		return request, invalid
	}
	return request, nil
}

func Reply(id string, data any, err error) Response {
	response := Response{CorrelationID: id, Status: "ok", Data: data}
	if err == nil {
		return response
	}

	response.Status = "error"
	response.Data = nil
	response.Error = &Error{Code: "internal_error", Message: "internal server error"}
	if domain, ok := errors.AsType[*model.Error](err); ok {
		response.Error = &Error{Code: domain.Code, Message: domain.Message}
	}
	return response
}
