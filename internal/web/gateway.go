package web

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"text/template"
)

var _ HttpGateway = (*HttpGatewayImpl)(nil)

//go:generate mockgen -destination mock/gateway_mock.go github.com/anibaldeboni/rapper/internal/web HttpGateway
type HttpGateway interface {
	Exec(variables map[string]string) (Response, error)
}

type HttpGatewayImpl struct {
	Token     string
	Method    string
	Client    HttpClient
	Templates struct {
		Url  *template.Template
		Body *template.Template
	}
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
func (this *HttpGatewayImpl) req(url string, body io.Reader, headers map[string]string) (Response, error) {
	if this.Method == http.MethodPost {
		return this.Client.Post(url, body, headers)
	}
	if this.Method == http.MethodPut {
		return this.Client.Put(url, body, headers)
	}
	return Response{}, errors.New("method not supported")
}

// Exec executes the request with the given variables to fill the body and url templates.
func (this *HttpGatewayImpl) Exec(variables map[string]string) (Response, error) {
	header := map[string]string{"Authorization": "Bearer " + this.Token}
	uri := RenderTemplate(this.Templates.Url, variables).String()
	body := RenderTemplate(this.Templates.Body, variables)

	return this.req(uri, body, header)
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
