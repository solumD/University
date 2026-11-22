package model

import "encoding/json"

type Command struct {
	ID      string          `json:"id"`
	Version string          `json:"version"`
	Action  string          `json:"action"`
	Data    json.RawMessage `json:"data"`
}

type Page[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func Protocol(code, message string) error {
	err := &Error{Code: code, Message: message}
	return err
}
