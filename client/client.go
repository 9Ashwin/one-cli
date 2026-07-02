// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

// Package client provides the HTTP client abstraction used by one-cli runtimes.
package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Request describes an API call to be made.
type Request struct {
	Method      string
	Path        string
	Query       map[string]string
	Headers     map[string]string
	Body        []byte
	AccessToken string
}

// Response describes the result of an API call.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// APIClient executes API requests.
type APIClient interface {
	Do(ctx context.Context, req Request) (*Response, error)
}

// HTTP is an APIClient backed by net/http.
type HTTP struct {
	BaseURL string
	Client  *http.Client
}

// Do implements APIClient.
func (c *HTTP) Do(ctx context.Context, req Request) (*Response, error) {
	url := strings.TrimRight(c.BaseURL, "/") + "/" + strings.TrimLeft(req.Path, "/")
	var body io.Reader
	if len(req.Body) > 0 {
		body = strings.NewReader(string(req.Body))
	}
	hreq, err := http.NewRequestWithContext(ctx, req.Method, url, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if req.AccessToken != "" {
		hreq.Header.Set("Authorization", "Bearer "+req.AccessToken)
	}
	for k, v := range req.Headers {
		hreq.Header.Set(k, v)
	}
	q := hreq.URL.Query()
	for k, v := range req.Query {
		q.Set(k, v)
	}
	hreq.URL.RawQuery = q.Encode()

	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	hresp, err := client.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer hresp.Body.Close()
	data, err := io.ReadAll(hresp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return &Response{
		StatusCode: hresp.StatusCode,
		Header:     hresp.Header,
		Body:       data,
	}, nil
}

// Fake is an in-memory APIClient for tests.
type Fake struct {
	Response *Response
	Err      error
	Requests []Request
}

// Do implements APIClient.
func (f *Fake) Do(_ context.Context, req Request) (*Response, error) {
	f.Requests = append(f.Requests, req)
	if f.Err != nil {
		return nil, f.Err
	}
	if f.Response != nil {
		return f.Response, nil
	}
	return &Response{StatusCode: 200, Body: []byte("{}")}, nil
}
