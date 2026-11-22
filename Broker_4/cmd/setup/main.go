package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"

	"gitlab.ozon.ru/university/broker-4/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "setup:", err)
		os.Exit(1)
	}
}

func run() (err error) {
	path := "config.local.json"
	if override := os.Getenv("CONFIG_PATH"); override != "" {
		path = override
	}
	u := url.URL{Scheme: "amqp", Host: "127.0.0.1:5672", Path: "/", User: url.UserPassword("broker4", rand.Text())}
	cfg := config.Config{RabbitURL: u.String(), APIKey: rand.Text(), Database: "data/broker.db", RateLimit: 60, LogLevel: "info"}
	body, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		fmt.Println("local configuration already exists; kept unchanged")
		return nil
	}
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	if _, err = file.Write(append(body, '\n')); err != nil {
		return err
	}
	fmt.Println("local configuration created; generated credentials are stored only in the local file")
	return nil
}
