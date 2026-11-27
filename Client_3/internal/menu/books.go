package menu

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"gitlab.ozon.ru/university/client-3/pkg/library"
)

func bookRow(b library.Book) string {
	line := fmt.Sprintf("Книга #%d · %s · %d · автор #%d", b.ID, b.Title, b.Year, b.AuthorID)

	return line
}

func (m *Menu) books(ctx context.Context) error {
	err := browse(ctx, m, m.client.Books, bookRow)

	return err
}

func (m *Menu) selectBook(ctx context.Context) (library.Book, error) {
	id, err := m.id("ID книги")
	if err != nil {
		return library.Book{}, err
	}

	book, err := m.client.Book(ctx, id)

	return book, err
}

func (m *Menu) book(ctx context.Context) error {
	book, err := m.selectBook(ctx)
	if err != nil {
		return err
	}

	fmt.Println(bookRow(book))

	return nil
}

func (m *Menu) bookFields(b library.Book) (library.Book, error) {
	var year, author string
	if b.ID != 0 {
		year = strconv.Itoa(b.Year)
		author = strconv.Itoa(b.AuthorID)
	}

	title, err := m.text("Название", b.Title)
	if err != nil {
		return b, err
	}

	published, err := m.number("Год издания", year, math.MinInt, math.MaxInt)
	if err != nil {
		return b, err
	}

	authorID, err := m.number("ID автора", author, 1, math.MaxInt)
	if err != nil {
		return b, err
	}

	b.Title = title
	b.Year = published
	b.AuthorID = authorID

	return b, nil
}

func (m *Menu) addBook(ctx context.Context) error {
	book, err := m.bookFields(library.Book{})
	if err != nil {
		return err
	}

	result, err := m.client.CreateBook(ctx, book)
	if err != nil {
		return err
	}

	fmt.Println("Добавлена:", bookRow(result))

	return nil
}

func (m *Menu) editBook(ctx context.Context) error {
	book, err := m.selectBook(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Enter сохраняет текущее значение.")
	book, err = m.bookFields(book)
	if err != nil {
		return err
	}

	result, err := m.client.UpdateBook(ctx, book)
	if err != nil {
		return err
	}

	fmt.Println("Обновлена:", bookRow(result))

	return nil
}

func (m *Menu) deleteBook(ctx context.Context) error {
	id, err := m.id("ID книги")
	if err != nil {
		return err
	}

	line, err := m.line(fmt.Sprintf("Удалить книгу #%d? [д/Н]: ", id))
	if err != nil {
		return err
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "д", "да", "y", "yes":
	default:
		fmt.Println("Удаление отменено.")

		return nil
	}

	err = m.client.DeleteBook(ctx, id)
	if err != nil {
		return err
	}

	fmt.Printf("Книга #%d удалена.\n", id)

	return nil
}
