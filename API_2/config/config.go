package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

type Config struct {
	Address        string `json:"http_addr"`
	Database       string `json:"db_path"`
	APIKey         string `json:"api_key"`
	InternalAPIKey string `json:"internal_api_key"`
	RateLimit      int    `json:"rate_limit"`
	LogLevel       string `json:"log_level"`
}

func Load() (Config, error) {
	cfg := Config{
		Address:   "127.0.0.1:8080",
		Database:  "data/api.db",
		RateLimit: 60,
		LogLevel:  "info",
	}

	if err := loadFile(&cfg); err != nil {
		return Config{}, err
	}

	cfg.Address = value("HTTP_ADDR", cfg.Address)
	cfg.Database = value("DB_PATH", cfg.Database)
	cfg.APIKey = value("API_KEY", cfg.APIKey)
	cfg.InternalAPIKey = value("INTERNAL_API_KEY", cfg.InternalAPIKey)
	cfg.LogLevel = value("LOG_LEVEL", cfg.LogLevel)
	if cfg.APIKey == "" {
		validationErr := errors.New("set api_key in config.local.json or set API_KEY")

		return Config{}, validationErr
	}

	if cfg.Address == "" || cfg.Database == "" {
		validationErr := errors.New("http_addr and db_path must not be empty")

		return Config{}, validationErr
	}

	if cfg.InternalAPIKey != "" && cfg.InternalAPIKey == cfg.APIKey {
		validationErr := errors.New("internal_api_key must differ from api_key")

		return Config{}, validationErr
	}

	limit, err := strconv.Atoi(value("RATE_LIMIT", strconv.Itoa(cfg.RateLimit)))
	if err != nil || limit < 1 || limit > 1000000 {
		validationErr := errors.New("rate_limit must be an integer between 1 and 1000000")

		return Config{}, validationErr
	}

	cfg.RateLimit = limit

	return cfg, nil
}

func loadFile(cfg *Config) error {
	data, err := os.ReadFile("config.local.json")
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		validationErr := fmt.Errorf("read config.local.json: %w", err)

		return validationErr
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err = decoder.Decode(cfg); err != nil {
		validationErr := fmt.Errorf("decode config.local.json: %w", err)

		return validationErr
	}

	var extra any

	if err = decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		validationErr := errors.New("config.local.json must contain exactly one JSON object")

		return validationErr
	}

	return nil
}

func value(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}

	return fallback
}
