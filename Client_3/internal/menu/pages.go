package menu

import (
	"context"
	"fmt"
	"strings"

	"gitlab.ozon.ru/university/client-3/pkg/library"
)

func browse[T any](ctx context.Context, m *Menu, list func(context.Context, int, int) (library.Page[T], error), row func(T) string) error {
	limit, err := m.number("Записей на странице", "10", 1, 100)
	if err != nil {
		return err
	}

	offset := 0
	for {
		page, err := list(ctx, limit, offset)
		if err != nil {
			return err
		}

		if len(page.Items) == 0 {
			fmt.Printf("Страница пуста. Всего записей: %d.\n", page.Total)
		} else {
			fmt.Printf("Записи %d–%d из %d\n", offset+1, offset+len(page.Items), page.Total)
			for _, item := range page.Items {
				fmt.Println(row(item))
			}
		}

		if offset == 0 && limit >= page.Total {
			return nil
		}

		next, more, err := m.nextPage(offset, limit, page.Total)
		if err != nil {
			return err
		}
		if !more {
			return nil
		}

		offset = next
	}
}

func (m *Menu) nextPage(offset, limit, total int) (int, bool, error) {
	for {
		line, err := m.line("n — вперёд, p — назад, Enter — в меню: ")
		if err != nil {
			return 0, false, err
		}

		switch strings.ToLower(strings.TrimSpace(line)) {
		case "":
			return 0, false, nil
		case "n":
			if offset+limit < total {
				return offset + limit, true, nil
			}

			fmt.Println("Это последняя страница.")
		case "p":
			if offset > 0 {
				previous := max(0, offset-limit)

				return previous, true, nil
			}

			fmt.Println("Это первая страница.")
		default:
			fmt.Println("Введите n, p или нажмите Enter.")
		}
	}
}
