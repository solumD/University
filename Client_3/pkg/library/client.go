package library

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const DefaultURL = "http://127.0.0.1:8000"

type Client struct {
	http      *resty.Client
	token     string
	remaining string
	retryAt   time.Time
}

func New(baseURL string) (*Client, error) {
	address, err := url.Parse(baseURL)
	if err != nil || (address.Scheme != "http" && address.Scheme != "https") || address.Hostname() == "" ||
		address.User != nil || address.RawQuery != "" || address.Fragment != "" {
		invalid := errors.New("LIBRARY_API_URL должен содержать HTTP(S)-адрес сервера без учётных данных, параметров и фрагмента")

		return nil, invalid
	}

	transport := resty.New().
		SetBaseURL(strings.TrimRight(baseURL, "/")).
		SetTimeout(5*time.Second).
		SetHeader("Accept", "application/json").
		SetRedirectPolicy(resty.NoRedirectPolicy())

	client := &Client{http: transport}

	return client, nil
}

func (c *Client) LoggedIn() bool {
	return c.token != ""
}

func (c *Client) Remaining() string {
	return c.remaining
}

func (c *Client) Login(ctx context.Context, username, password string) error {
	form := map[string]string{"username": username, "password": password}
	request := c.http.R().SetFormData(form)

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	err := c.send(ctx, request, http.MethodPost, "/token", &result)
	if err != nil {
		return err
	}

	if result.AccessToken == "" || !strings.EqualFold(result.TokenType, "bearer") {
		invalid := errors.New("сервер вернул некорректный ответ при входе")

		return invalid
	}

	c.token = result.AccessToken

	return nil
}

func (c *Client) send(ctx context.Context, request *resty.Request, method, path string, result any) error {
	if wait := time.Until(c.retryAt); wait > 0 {
		limitErr := &APIError{Status: http.StatusTooManyRequests, RetryAfter: wait}

		return limitErr
	}

	response, err := request.SetContext(ctx).Execute(method, path)
	if err != nil {
		failure := fmt.Errorf("запрос к API: %w", err)

		return failure
	}

	c.remaining = response.Header().Get("X-Limit-Remaining")

	if !response.IsSuccess() {
		failure := readError(response)
		if failure.Status == http.StatusUnauthorized {
			c.token = ""
		}

		if failure.Status == http.StatusTooManyRequests && failure.RetryAfter > 0 {
			c.retryAt = time.Now().Add(failure.RetryAfter)
		}

		return failure
	}

	if result == nil {
		return nil
	}

	if err = json.Unmarshal(response.Body(), result); err != nil {
		failure := fmt.Errorf("не удалось прочитать JSON в ответе API: %w", err)

		return failure
	}

	return nil
}
