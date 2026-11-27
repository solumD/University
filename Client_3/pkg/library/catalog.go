package library

import (
	"context"
	"errors"
	"net/http"
	"strconv"
)

const (
	booksPath   = "/api/v1/books"
	authorsPath = "/api/v1/authors"
	readersPath = "/api/v1/users"
	loansPath   = "/api/v1/borrow"
)

func resourcePath(collection string, id int) string {
	path := collection + "/" + strconv.Itoa(id)

	return path
}

func page[T any](ctx context.Context, c *Client, path string, limit, offset int, fields string) (Page[T], error) {
	request := c.http.R().
		SetQueryParam("limit", strconv.Itoa(limit)).
		SetQueryParam("offset", strconv.Itoa(offset))

	if fields != "" {
		request.SetQueryParam("include", fields)
	}

	var result Page[T]
	err := c.send(ctx, request, http.MethodGet, path, &result)
	if err != nil {
		return Page[T]{}, err
	}

	if result.Items == nil || result.Limit != limit || result.Offset != offset || result.Total < 0 {
		invalid := errors.New("неожиданный формат страницы в ответе API")

		return Page[T]{}, invalid
	}

	return result, nil
}

func (c *Client) Books(ctx context.Context, limit, offset int) (Page[Book], error) {
	result, err := page[Book](ctx, c, booksPath, limit, offset, "")

	return result, err
}

func (c *Client) Book(ctx context.Context, id int) (Book, error) {
	var book Book
	err := c.send(ctx, c.http.R(), http.MethodGet, resourcePath(booksPath, id), &book)

	return book, err
}

func (c *Client) CreateBook(ctx context.Context, book Book) (Book, error) {
	var result Book
	request := c.http.R().SetBody(book)
	err := c.send(ctx, request, http.MethodPost, booksPath, &result)

	return result, err
}

func (c *Client) UpdateBook(ctx context.Context, book Book) (Book, error) {
	path := resourcePath(booksPath, book.ID)
	book.ID = 0

	var result Book
	request := c.http.R().SetBody(book)
	err := c.send(ctx, request, http.MethodPut, path, &result)

	return result, err
}

func (c *Client) DeleteBook(ctx context.Context, id int) error {
	err := c.send(ctx, c.http.R(), http.MethodDelete, resourcePath(booksPath, id), nil)

	return err
}

func (c *Client) Authors(ctx context.Context, limit, offset int) (Page[Author], error) {
	result, err := page[Author](ctx, c, authorsPath, limit, offset, "")

	return result, err
}

func (c *Client) CreateAuthor(ctx context.Context, author Author) (Author, error) {
	var result Author
	request := c.http.R().SetBody(author)
	err := c.send(ctx, request, http.MethodPost, authorsPath, &result)

	return result, err
}

func (c *Client) Readers(ctx context.Context, limit, offset int) (Page[Reader], error) {
	result, err := page[Reader](ctx, c, readersPath, limit, offset, "id,username")

	return result, err
}

func (c *Client) CreateReader(ctx context.Context, username, password string) (Reader, error) {
	var result Reader
	request := c.http.R().SetBody(map[string]string{"username": username, "password": password})
	err := c.send(ctx, request, http.MethodPost, readersPath, &result)

	return result, err
}
