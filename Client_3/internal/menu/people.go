package menu

import (
	"context"
	"fmt"

	"gitlab.ozon.ru/university/client-3/pkg/library"
)

func authorRow(a library.Author) string {
	line := fmt.Sprintf("Автор #%d · %s · %s", a.ID, a.Name, a.Country)

	return line
}

func (m *Menu) authors(ctx context.Context) error {
	err := browse(ctx, m, m.client.Authors, authorRow)

	return err
}

func (m *Menu) addAuthor(ctx context.Context) error {
	name, err := m.text("Имя автора", "")
	if err != nil {
		return err
	}

	country, err := m.text("Страна", "")
	if err != nil {
		return err
	}

	author, err := m.client.CreateAuthor(ctx, library.Author{Name: name, Country: country})
	if err != nil {
		return err
	}

	fmt.Println(authorRow(author))

	return nil
}

func readerRow(r library.Reader) string {
	line := fmt.Sprintf("Читатель #%d · %s", r.ID, r.Username)

	return line
}

func (m *Menu) readers(ctx context.Context) error {
	err := browse(ctx, m, m.client.Readers, readerRow)

	return err
}

func (m *Menu) credentials() (string, string, error) {
	name, err := m.text("Логин", "")
	if err != nil {
		return "", "", err
	}

	password, err := m.password()

	return name, password, err
}

func (m *Menu) addReader(ctx context.Context) error {
	name, password, err := m.credentials()
	if err != nil {
		return err
	}

	reader, err := m.client.CreateReader(ctx, name, password)
	if err != nil {
		return err
	}

	fmt.Println(readerRow(reader))

	return nil
}

func (m *Menu) login(ctx context.Context) error {
	name, password, err := m.credentials()
	if err != nil {
		return err
	}

	err = m.client.Login(ctx, name, password)
	if err != nil {
		return err
	}

	if m.account != name {
		m.last = nil
	}
	m.account = name
	fmt.Printf("Вход выполнен: %s.\n", name)

	return nil
}
