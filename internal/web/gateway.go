package web

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"text/template"
)

var _ HttpGateway = (*HttpGatewayImpl)(nil)

//go:generate mockgen -destination mock/gateway_mock.go github.com/anibaldeboni/rapper/internal/web HttpGateway
type HttpGateway interface {
	Exec(ctx context.Context, variables map[string]string) (Response, error)
	UpdateConfig(method, urlTemplate, bodyTemplate string, headers map[string]string) error
}

type HttpGatewayImpl struct {
	Client       HttpClient
	urlTemplate  *template.Template
	bodyTemplate *template.Template
	headers      map[string]string // Flexible headers (Authorization, Cookie, etc)
	method       string
	mu           sync.RWMutex // Protects against concurrent access during hot-reload
}

// NewHttpGateway creates a new HttpGateway with flexible headers.
func NewHttpGateway(method, urlTemplate, bodyTemplate string, headers map[string]string) (HttpGateway, error) {
	urlTmpl, err := NewTemplate("url", urlTemplate)
	if err != nil {
		return nil, fmt.Errorf("invalid URL template: %w", err)
	}

	bodyTmpl, err := NewTemplate("body", bodyTemplate)
	if err != nil {
		return nil, fmt.Errorf("invalid body template: %w", err)
	}

	return &HttpGatewayImpl{
		method:       method,
		urlTemplate:  urlTmpl,
		bodyTemplate: bodyTmpl,
		headers:      headers,
		Client:       NewHttpClient(),
	}, nil
}

// NewHttpGatewayLegacy creates a new HttpGateway using legacy token-based auth.
// DEPRECATED: Use NewHttpGateway with headers map instead.
func NewHttpGatewayLegacy(token, method, urlTemplate, bodyTemplate string) HttpGateway {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}

	gateway, err := NewHttpGateway(method, urlTemplate, bodyTemplate, headers)
	if err != nil {
		// Fallback to basic implementation if templates fail
		return &HttpGatewayImpl{
			method:  method,
			headers: headers,
			Client:  NewHttpClient(),
		}
	}

	return gateway
}
func (hg *HttpGatewayImpl) req(ctx context.Context, url string, body io.Reader, headers map[string]string) (Response, error) {
	hg.mu.RLock()
	method := hg.method
	hg.mu.RUnlock()

	switch method {
	case http.MethodGet:
		return hg.Client.Get(ctx, url, headers)
	case http.MethodPost:
		return hg.Client.Post(ctx, url, body, headers)
	case http.MethodPut:
		return hg.Client.Put(ctx, url, body, headers)
	case http.MethodDelete:
		return hg.Client.Delete(ctx, url, headers)
	case http.MethodPatch:
		return hg.Client.Patch(ctx, url, body, headers)
	default:
		return Response{}, errors.New("method not supported: " + method)
	}
}

// Exec executes the request with the given variables to fill the body and url templates.
// It supports template rendering for both URL, body, and header values.
func (hg *HttpGatewayImpl) Exec(ctx context.Context, variables map[string]string) (Response, error) {
	hg.mu.RLock()
	defer hg.mu.RUnlock()

	// Render URL template
	uri := RenderTemplate(hg.urlTemplate, variables).String()

	// Render body template
	body := RenderTemplate(hg.bodyTemplate, variables)

	// Render headers (supports templates in header values)
	headers := make(map[string]string)
	for key, value := range hg.headers {
		// Try to render as template (for dynamic headers like Authorization: Bearer {{.token}})
		rendered, err := renderString(value, variables)
		if err != nil {
			// If rendering fails, use literal value
			headers[key] = value
		} else {
			headers[key] = rendered
		}
	}

	return hg.req(ctx, uri, body, headers)
}

// UpdateConfig updates the gateway configuration and templates at runtime (hot-reload)
func (hg *HttpGatewayImpl) UpdateConfig(method, urlTemplate, bodyTemplate string, headers map[string]string) error {
	hg.mu.Lock()
	defer hg.mu.Unlock()

	// Parse new templates
	urlTmpl, err := NewTemplate("url", urlTemplate)
	if err != nil {
		return fmt.Errorf("invalid URL template: %w", err)
	}

	bodyTmpl, err := NewTemplate("body", bodyTemplate)
	if err != nil {
		return fmt.Errorf("invalid body template: %w", err)
	}

	// Update all fields atomically
	hg.method = method
	hg.urlTemplate = urlTmpl
	hg.bodyTemplate = bodyTmpl
	hg.headers = headers

	return nil
}

// NewTemplate creates a new template from a string
func NewTemplate(name string, templ string) (*template.Template, error) {
	return template.New(name).Parse(templ)
}

// RenderTemplate renders a template with the given variables
func RenderTemplate[T map[string]string](t *template.Template, variables T) *bytes.Buffer {
	var result string
	buf := bytes.NewBufferString(result)
	_ = t.Execute(buf, variables)

	return buf
}

// renderString renders a string template with variables
func renderString(templateStr string, variables map[string]string) (string, error) {
	tmpl, err := template.New("").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, variables)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
