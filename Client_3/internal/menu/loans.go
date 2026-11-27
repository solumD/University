package menu

import (
	"context"
	"crypto/rand"
	"fmt"

	"gitlab.ozon.ru/university/client-3/pkg/library"
)

func printLoan(loan library.Loan) {
	fmt.Printf("Выдача #%d · книга #%d · читатель #%d · статус: %s\n",
		loan.ID, loan.BookID, loan.UserID, loan.Status)
	if loan.BorrowDate != "" {
		fmt.Println("Дата выдачи:", loan.BorrowDate)
	}
	if loan.ReturnDate != "" {
		fmt.Println("Дата возврата:", loan.ReturnDate)
	}
}

func (m *Menu) issue(ctx context.Context) error {
	if !m.client.LoggedIn() {
		return library.ErrLoginRequired
	}

	book, err := m.id("ID книги")
	if err != nil {
		return err
	}

	user, err := m.id("ID читателя")
	if err != nil {
		return err
	}

	m.last = &issueAttempt{
		data: library.LoanRequest{BookID: book, UserID: user},
		key:  rand.Text(),
	}

	err = m.repeat(ctx)

	return err
}

func (m *Menu) repeat(ctx context.Context) error {
	if !m.client.LoggedIn() {
		return library.ErrLoginRequired
	}

	if m.last == nil {
		fmt.Println("В этом сеансе ещё не оформляли выдачу.")

		return nil
	}

	fmt.Println("Ключ операции:", m.last.key)
	loan, err := m.client.Issue(ctx, m.last.data, m.last.key)
	if err != nil {
		return err
	}

	printLoan(loan)

	return nil
}

func (m *Menu) loan(ctx context.Context) error {
	if !m.client.LoggedIn() {
		return library.ErrLoginRequired
	}

	id, err := m.id("ID выдачи")
	if err != nil {
		return err
	}

	loan, err := m.client.Loan(ctx, id)
	if err != nil {
		return err
	}

	printLoan(loan)

	return nil
}

func (m *Menu) returnBook(ctx context.Context) error {
	if !m.client.LoggedIn() {
		return library.ErrLoginRequired
	}

	id, err := m.id("ID выдачи")
	if err != nil {
		return err
	}

	loan, err := m.client.Return(ctx, id)
	if err != nil {
		return err
	}

	printLoan(loan)

	return nil
}
