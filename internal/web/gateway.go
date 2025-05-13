package web

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"text/template"
)

var _ HttpGateway = (*HttpGatewayImpl)(nil)

//go:generate mockgen -destination mock/gateway_mock.go github.com/anibaldeboni/rapper/internal/web HttpGateway
type HttpGateway interface {
	Exec(ctx context.Context, variables map[string]string) (Response, error)
}

type HttpGatewayImpl struct {
	Client    HttpClient
	Templates struct {
		Url  *template.Template
		Body *template.Template
	}
	Token  string
	Method string
}

// NewHttpGateway creates a new HttpGateway.
func NewHttpGateway(token, method, urlTemplate, bodyTemplate string) HttpGateway {
	return &HttpGatewayImpl{
		Token:  token,
		Method: method,
		Client: NewHttpClient(),
		Templates: struct {
			Url  *template.Template
			Body *template.Template
		}{
			NewTemplate("url", urlTemplate),
			NewTemplate("body", bodyTemplate),
		},
	}
}
func (hg *HttpGatewayImpl) req(ctx context.Context, url string, body io.Reader, headers map[string]string) (Response, error) {
	switch hg.Method {
	case http.MethodGet:
		return hg.Client.Get(ctx, url, headers)
	case http.MethodPost:
		return hg.Client.Post(ctx, url, body, headers)
	case http.MethodPut:
		return hg.Client.Put(ctx, url, body, headers)
	default:
		return Response{}, errors.New("method not supported: " + hg.Method)
	}
}

// Exec executes the request with the given variables to fill the body and url templates.
func (hg *HttpGatewayImpl) Exec(ctx context.Context, variables map[string]string) (Response, error) {
	header := map[string]string{"Authorization": "Bearer " + hg.Token}
	uri := RenderTemplate(hg.Templates.Url, variables).String()
	body := RenderTemplate(hg.Templates.Body, variables)

	return hg.req(ctx, uri, body, header)
}

func NewTemplate(name string, templ string) *template.Template {
	return template.Must(template.New(name).Parse(templ))
}

func RenderTemplate[T map[string]string](t *template.Template, variables T) *bytes.Buffer {
	var result string
	buf := bytes.NewBufferString(result)
	_ = t.Execute(buf, variables)

	return buf
}
