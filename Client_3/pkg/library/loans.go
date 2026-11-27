package library

import (
	"context"
	"net/http"

	"github.com/go-resty/resty/v2"
)

func (c *Client) authenticated() (*resty.Request, error) {
	if !c.LoggedIn() {
		return nil, ErrLoginRequired
	}

	request := c.http.R().SetAuthToken(c.token)

	return request, nil
}

func (c *Client) Issue(ctx context.Context, loan LoanRequest, key string) (Loan, error) {
	request, err := c.authenticated()
	if err != nil {
		return Loan{}, err
	}

	request.SetHeader("Idempotency-Key", key).SetBody(loan)

	var result Loan
	err = c.send(ctx, request, http.MethodPost, loansPath, &result)

	return result, err
}

func (c *Client) Loan(ctx context.Context, id int) (Loan, error) {
	request, err := c.authenticated()
	if err != nil {
		return Loan{}, err
	}

	var result Loan
	err = c.send(ctx, request, http.MethodGet, resourcePath(loansPath, id), &result)

	return result, err
}

func (c *Client) Return(ctx context.Context, id int) (Loan, error) {
	request, err := c.authenticated()
	if err != nil {
		return Loan{}, err
	}

	path := resourcePath(loansPath, id) + "/return"

	var result Loan
	err = c.send(ctx, request, http.MethodPut, path, &result)

	return result, err
}
