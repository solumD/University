package menu

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"gitlab.ozon.ru/university/client-3/pkg/library"
)

const choices = `
КНИГИ        1  Список      2  Найти по ID     3  Добавить     4  Изменить     5  Удалить
АВТОРЫ       6  Список      7  Добавить
ЧИТАТЕЛИ     8  Список      9  Создать        10  Войти
ВЫДАЧИ      11  Выдать     12  Повторить      13  Найти по ID  14  Вернуть

?  Показать меню            0  Выход`

type issueAttempt struct {
	data library.LoanRequest
	key  string
}

type Menu struct {
	client  *library.Client
	input   *bufio.Reader
	account string
	last    *issueAttempt
}

func New(client *library.Client) *Menu {
	m := &Menu{client: client, input: bufio.NewReader(os.Stdin)}

	return m
}

func (m *Menu) Run(ctx context.Context) error {
	actions := map[string]func(context.Context) error{
		"1": m.books, "2": m.book, "3": m.addBook, "4": m.editBook, "5": m.deleteBook,
		"6": m.authors, "7": m.addAuthor,
		"8": m.readers, "9": m.addReader, "10": m.login,
		"11": m.issue, "12": m.repeat, "13": m.loan, "14": m.returnBook,
	}

	fmt.Println(choices)

	for {
		account := "гость"
		if m.client.LoggedIn() {
			account = m.account
		}

		fmt.Printf("\nСеанс: %s", account)
		if remaining := m.client.Remaining(); remaining != "" {
			fmt.Printf(" · осталось запросов: %s", remaining)
		}

		command, err := m.line("\nДействие (? — меню): ")
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		command = strings.TrimSpace(command)
		switch command {
		case "0":
			return nil
		case "?":
			fmt.Println(choices)
			continue
		}

		action, ok := actions[command]
		if !ok {
			fmt.Println("Введите номер действия или ? для просмотра меню.")
			continue
		}

		err = action(ctx)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("Время ожидания истекло. Если прервалась выдача, выберите 12 для повтора с тем же ключом.")
			continue
		}
		if err != nil {
			fmt.Println("Ошибка:", err)
		}
	}
}
