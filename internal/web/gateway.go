package web

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"text/template"
)

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
	return HttpGatewayImpl{
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
func (hg *HttpGatewayImpl) req(url string, body io.Reader, headers map[string]string) (Response, error) {
	if hg.Method == http.MethodPost {
		return hg.Client.Post(url, body, headers)
	}
	if hg.Method == http.MethodPut {
		return hg.Client.Put(url, body, headers)
	}
	return Response{}, errors.New("method not supported")
}

// Exec executes the request with the given variables to fill the body and url templates.
func (hg HttpGatewayImpl) Exec(variables map[string]string) (Response, error) {
	header := map[string]string{"Authorization": "Bearer " + hg.Token}
	uri := RenderTemplate(hg.Templates.Url, variables).String()
	body := RenderTemplate(hg.Templates.Body, variables)

	return hg.req(uri, body, header)
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
