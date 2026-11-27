package main

import (
	"context"
	"fmt"
	"os"

	"gitlab.ozon.ru/university/client-3/internal/menu"
	"gitlab.ozon.ru/university/client-3/pkg/library"
)

func main() {
	address := os.Getenv("LIBRARY_API_URL")
	if address == "" {
		address = library.DefaultURL
	}

	client, err := library.New(address)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Библиотека ·", address)
	screen := menu.New(client)

	err = screen.Run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
