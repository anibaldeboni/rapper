package web

import (
	"errors"
	"io"
	"net/http"
	"rapper/files"
	"text/template"
)

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

func NewHttpGateway(token, method, urlTemplate, bodyTemplate string) HttpGateway {
	return &HttpGatewayImpl{
		Token:  token,
		Method: method,
		Client: NewHttpClient(),
		Templates: struct {
			Url  *template.Template
			Body *template.Template
		}{
			files.NewTemplate("url", urlTemplate),
			files.NewTemplate("body", bodyTemplate),
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
func (hg *HttpGatewayImpl) Exec(variables map[string]string) (Response, error) {
	header := map[string]string{"Authorization": "Bearer " + hg.Token}
	uri := files.RenderTemplate(hg.Templates.Url, variables).String()
	body := files.RenderTemplate(hg.Templates.Body, variables)

	return hg.req(uri, body, header)
}
