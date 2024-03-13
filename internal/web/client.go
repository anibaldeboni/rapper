package web

import (
	"io"
	"net/http"
)

var _ HttpClient = (*httpClientImpl)(nil)

type Response struct {
	URL        string
	StatusCode int
	Headers    http.Header
	Body       []byte
}

//go:generate mockgen -destination mock/client_mock.go github.com/anibaldeboni/rapper/internal/web HttpClient
type HttpClient interface {
	Put(url string, body io.Reader, headers map[string]string) (Response, error)
	Post(url string, body io.Reader, headers map[string]string) (Response, error)
	Get(url string, headers map[string]string) (Response, error)
}

type httpClientImpl struct{}

func NewHttpClient() HttpClient {
	return &httpClientImpl{}
}

func (httpClientImpl) Put(url string, body io.Reader, headers map[string]string) (Response, error) {
	headers = buildHeaders(headers)
	return request(http.MethodPut, url, headers, body)
}

func (httpClientImpl) Post(url string, body io.Reader, headers map[string]string) (Response, error) {
	headers = buildHeaders(headers)
	return request(http.MethodPost, url, headers, body)
}

func (httpClientImpl) Get(url string, headers map[string]string) (Response, error) {
	return request(http.MethodGet, url, headers, nil)
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

func request(method string, url string, headers map[string]string, body io.Reader) (Response, error) {
	req, err := http.NewRequest(method, url, body)
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
