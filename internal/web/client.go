package web

import (
	"context"
	"io"
	"maps"
	"net/http"
)

// Response represents an HTTP response. Method is captured by the
// client alongside URL so downstream consumers (logs.NewHTTPMessage)
// can render "METHOD URL status" without re-deriving the verb from
// the gateway config.
type Response struct {
	Headers    http.Header
	Method     string
	URL        string
	Body       []byte
	StatusCode int
}

// httpClientImpl handles raw HTTP operations.
// This is an internal implementation detail, not exposed as an interface.
type httpClientImpl struct{}

// newHttpClient creates a new HTTP client.
func newHttpClient() *httpClientImpl {
	return &httpClientImpl{}
}

// NewHttpClient creates a new HTTP client for external use (e.g., updates package).
// Returns the concrete type for direct use without interface abstraction.
func NewHttpClient() *httpClientImpl {
	return &httpClientImpl{}
}

func (c *httpClientImpl) Put(ctx context.Context, url string, body io.Reader, headers map[string]string) (Response, error) {
	headers = buildHeaders(headers)
	return request(ctx, http.MethodPut, url, headers, body)
}

func (c *httpClientImpl) Post(ctx context.Context, url string, body io.Reader, headers map[string]string) (Response, error) {
	headers = buildHeaders(headers)
	return request(ctx, http.MethodPost, url, headers, body)
}

func (c *httpClientImpl) Get(ctx context.Context, url string, headers map[string]string) (Response, error) {
	return request(ctx, http.MethodGet, url, headers, nil)
}

func (c *httpClientImpl) Delete(ctx context.Context, url string, headers map[string]string) (Response, error) {
	headers = buildHeaders(headers)
	return request(ctx, http.MethodDelete, url, headers, nil)
}

func (c *httpClientImpl) Patch(ctx context.Context, url string, body io.Reader, headers map[string]string) (Response, error) {
	headers = buildHeaders(headers)
	return request(ctx, http.MethodPatch, url, headers, body)
}

func addHeaders(headers map[string]string, request *http.Request) {
	for k, v := range headers {
		request.Header.Set(k, v)
	}
}
func buildHeaders(m2 map[string]string) map[string]string {
	m1 := map[string]string{"Content-Type": "application/json; charset=UTF-8"}
	maps.Copy(m1, m2)

	return m1
}

func request(ctx context.Context, method string, url string, headers map[string]string, body io.Reader) (Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return Response{Method: method, URL: url}, err
	}
	addHeaders(headers, req)
	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return Response{Method: method, URL: url}, err
	}
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	return Response{
		Method:     res.Request.Method,
		URL:        res.Request.URL.String(),
		StatusCode: res.StatusCode,
		Headers:    res.Header,
		Body:       resBody,
	}, nil
}
