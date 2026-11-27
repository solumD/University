package menu

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

func (m *Menu) line(prompt string) (string, error) {
	fmt.Print(prompt)
	line, err := m.input.ReadString('\n')
	if errors.Is(err, io.EOF) && line != "" {
		err = nil
	}

	line = strings.TrimRight(line, "\r\n")

	return line, err
}

func (m *Menu) text(label, fallback string) (string, error) {
	if fallback != "" {
		label += " [" + fallback + "]"
	}

	for {
		line, err := m.line(label + ": ")
		if err != nil {
			return "", err
		}

		value := strings.TrimSpace(line)
		if value == "" {
			value = fallback
		}
		if value != "" {
			return value, nil
		}

		fmt.Println("Введите значение.")
	}
}

func (m *Menu) number(label, fallback string, minimum, maximum int) (int, error) {
	for {
		value, err := m.text(label, fallback)
		if err != nil {
			return 0, err
		}

		n, err := strconv.Atoi(value)
		if err == nil && minimum <= n && n <= maximum {
			return n, nil
		}

		switch {
		case minimum == 1 && maximum == math.MaxInt:
			fmt.Println("Введите положительный целый ID.")
		case minimum == math.MinInt && maximum == math.MaxInt:
			fmt.Println("Введите целое число.")
		default:
			fmt.Printf("Введите целое число от %d до %d.\n", minimum, maximum)
		}
	}
}

func (m *Menu) id(label string) (int, error) {
	n, err := m.number(label, "", 1, math.MaxInt)

	return n, err
}

func (m *Menu) password() (string, error) {
	fd := int(os.Stdin.Fd())

	for {
		var value string
		var err error

		if term.IsTerminal(fd) {
			fmt.Print("Пароль: ")
			var data []byte
			data, err = term.ReadPassword(fd)
			fmt.Println()
			value = string(data)
		} else {
			value, err = m.line("Пароль: ")
		}

		if err != nil {
			return "", err
		}
		if value != "" {
			return value, nil
		}

		fmt.Println("Введите пароль.")
	}
}
