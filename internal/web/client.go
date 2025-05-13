package web

import (
	"context"
	"io"
	"net/http"
)

var _ HttpClient = (*httpClientImpl)(nil)

type Response struct {
	Headers    http.Header
	URL        string
	Body       []byte
	StatusCode int
}

//go:generate mockgen -destination mock/client_mock.go github.com/anibaldeboni/rapper/internal/web HttpClient
type HttpClient interface {
	Put(ctx context.Context, url string, body io.Reader, headers map[string]string) (Response, error)
	Post(ctx context.Context, url string, body io.Reader, headers map[string]string) (Response, error)
	Get(ctx context.Context, url string, headers map[string]string) (Response, error)
}

type httpClientImpl struct{}

func NewHttpClient() HttpClient {
	return &httpClientImpl{}
}

func (httpClientImpl) Put(ctx context.Context, url string, body io.Reader, headers map[string]string) (Response, error) {
	headers = buildHeaders(headers)
	return request(ctx, http.MethodPut, url, headers, body)
}

func (httpClientImpl) Post(ctx context.Context, url string, body io.Reader, headers map[string]string) (Response, error) {
	headers = buildHeaders(headers)
	return request(ctx, http.MethodPost, url, headers, body)
}

func (httpClientImpl) Get(ctx context.Context, url string, headers map[string]string) (Response, error) {
	return request(ctx, http.MethodGet, url, headers, nil)
}

func addHeaders(headers map[string]string, request *http.Request) {
	for k, v := range headers {
		request.Header.Set(k, v)
	}
}
func buildHeaders(m2 map[string]string) map[string]string {
	m1 := map[string]string{"Content-Type": "application/json; charset=UTF-8"}
	for k, v := range m2 {
		m1[k] = v
	}

	return m1
}

func request(ctx context.Context, method string, url string, headers map[string]string, body io.Reader) (Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return Response{}, err
	}
	addHeaders(headers, req)
	client := &http.Client{}

	return buildResponse(client.Do(req))
}

func buildResponse(response *http.Response, err error) (Response, error) {
	if err != nil {
		return Response{}, err
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)

	return Response{
		URL:        response.Request.URL.String(),
		StatusCode: response.StatusCode,
		Headers:    response.Header,
		Body:       body,
	}, nil
}
