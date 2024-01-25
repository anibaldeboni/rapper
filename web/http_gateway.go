package web

import (
	"rapper/files"
	"text/template"
)

type HttpGateway interface {
	Exec(variables map[string]string) (Response, error)
}
type requestFuncMap map[string]RequestFunc

type httpGatewayImpl struct {
	token     string
	method    string
	client    HttpClient
	templates struct {
		url  *template.Template
		body *template.Template
	}
}

func NewHttpGateway(token, method, urlTemplate, bodyTemplate string) HttpGateway {
	return &httpGatewayImpl{
		token:  token,
		method: method,
		client: NewHttpClient(),
		templates: struct {
			url  *template.Template
			body *template.Template
		}{
			files.NewTemplate("url", urlTemplate),
			files.NewTemplate("body", bodyTemplate),
		},
	}
}
func (hg *httpGatewayImpl) httpFunc(method string) RequestFunc {
	return requestFuncMap{
		"PUT":  hg.client.Put,
		"POST": hg.client.Post,
	}[method]
}

func (hg *httpGatewayImpl) Exec(variables map[string]string) (Response, error) {
	header := map[string]string{"Authorization": "Bearer " + hg.token}
	uri := files.RenderTemplate(hg.templates.url, variables).String()
	body := files.RenderTemplate(hg.templates.body, variables)

	return hg.httpFunc(hg.method)(uri, body, header)
}
