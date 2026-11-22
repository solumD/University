package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	RabbitURL string `json:"rabbitmq_url"`
	APIKey    string `json:"api_key"`
	Database  string `json:"db_path"`
	RateLimit int    `json:"rate_limit"`
	LogLevel  string `json:"log_level"`
}

const OperationTimeout = 5 * time.Second

func Load() (Config, error) {
	cfg := Config{Database: "data/broker.db", RateLimit: 60, LogLevel: "info"}
	path := value("CONFIG_PATH", "config.local.json")
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return cfg, err
	}

	if err == nil {
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(&cfg); err != nil {
			validationErr := errors.New("invalid local configuration")

			return cfg, validationErr
		}

		var extra any
		if err = decoder.Decode(&extra); !errors.Is(err, io.EOF) {
			validationErr := errors.New("configuration must contain one JSON object")

			return cfg, validationErr
		}
	}

	cfg.RabbitURL = value("RABBITMQ_URL", cfg.RabbitURL)
	cfg.APIKey = value("API_KEY", cfg.APIKey)
	cfg.Database = value("DB_PATH", cfg.Database)
	cfg.LogLevel = value("LOG_LEVEL", cfg.LogLevel)
	cfg.RateLimit, err = strconv.Atoi(value("RATE_LIMIT", strconv.Itoa(cfg.RateLimit)))
	if err != nil || cfg.RateLimit < 1 || cfg.RateLimit > 1000000 {
		validationErr := errors.New("rate_limit must be between 1 and 1000000")

		return cfg, validationErr
	}

	u, err := url.Parse(cfg.RabbitURL)
	if err != nil || u == nil || (u.Scheme != "amqp" && u.Scheme != "amqps") || u.Host == "" || u.User == nil {
		validationErr := errors.New("set rabbitmq_url with credentials in local configuration")

		return cfg, validationErr
	}

	password, present := u.User.Password()
	if !present || password == "" || u.User.Username() == "" || cfg.APIKey == "" || cfg.Database == "" {
		validationErr := errors.New("rabbitmq credentials, api_key and db_path must not be empty")

		return cfg, validationErr
	}

	return cfg, nil
}

func value(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}
