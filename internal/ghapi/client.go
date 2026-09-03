package ghapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiBase = "https://api.github.com"

type Client struct {
	Token string
	HTTP  *http.Client
}

func NewClient(token string) *Client {
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 && req.URL.Host != via[0].URL.Host {
				req.Header.Del("Authorization")
			}
			return nil
		},
	}
	return &Client{
		Token: token,
		HTTP:  httpClient,
	}
}

type APIError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("GitHub API error (%d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("GitHub API error (%d)", e.StatusCode)
}

func (c *Client) request(method, path string, body interface{}, out interface{}) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}

	url := path
	if len(path) < 8 || (path[:8] != "https://" && path[:7] != "http://") {
		url = apiBase + path
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
		var parsed struct {
			Message string `json:"message"`
		}
		if jsonErr := json.Unmarshal(respBody, &parsed); jsonErr == nil {
			apiErr.Message = parsed.Message
		}
		return apiErr
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Get(path string, out interface{}) error {
	return c.request(http.MethodGet, path, nil, out)
}

func (c *Client) Post(path string, body interface{}, out interface{}) error {
	return c.request(http.MethodPost, path, body, out)
}

func (c *Client) Put(path string, body interface{}, out interface{}) error {
	return c.request(http.MethodPut, path, body, out)
}

func (c *Client) Patch(path string, body interface{}, out interface{}) error {
	return c.request(http.MethodPatch, path, body, out)
}

func (c *Client) Delete(path string) error {
	return c.request(http.MethodDelete, path, nil, nil)
}

type User struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

func (c *Client) CurrentUser() (*User, error) {
	var u User
	if err := c.Get("/user", &u); err != nil {
		return nil, err
	}
	if u.Name == "" {
		u.Name = u.Login
	}
	return &u, nil
}
