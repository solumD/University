package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"time"

	"gitlab.ozon.ru/university/broker-4/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "broker:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 || (os.Args[1] != "up" && os.Args[1] != "down" && os.Args[1] != "queues") {
		validationErr := errors.New("usage: broker up|down|queues")

		return validationErr
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	u, err := url.Parse(cfg.RabbitURL)
	if err != nil {
		validationErr := errors.New("invalid broker URL")

		return validationErr
	}
	password, _ := u.User.Password()
	args := []string{"compose"}
	switch os.Args[1] {
	case "up":
		args = append(args, "up", "-d", "--wait", "--wait-timeout", "120")
	case "down":
		args = append(args, "down")
	case "queues":
		args = append(args, "exec", "-T", "--user", "rabbitmq", "rabbitmq", "rabbitmqctl", "list_queues", "name", "messages_ready", "messages_unacknowledged")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	executable := "docker"
	if standalone, lookupErr := exec.LookPath("docker-compose"); lookupErr == nil {
		executable = standalone
		args = args[1:]
	}
	command := exec.CommandContext(ctx, executable, args...)
	command.Env = append(os.Environ(), "BROKER_USER="+u.User.Username(), "BROKER_PASSWORD="+password)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	err = command.Run()
	return err
}
