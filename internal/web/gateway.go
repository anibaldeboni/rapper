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

// httpGatewayImpl implements the processor.HttpGateway interface.
// This is an internal implementation - the interface is defined by the client (processor package).
type httpGatewayImpl struct {
	client       *httpClientImpl
	urlTemplate  *template.Template
	bodyTemplate *template.Template
	headers      map[string]string // Flexible headers (Authorization, Cookie, etc)
	method       string
	mu           sync.RWMutex // Protects against concurrent access during hot-reload
}

// NewHttpGateway creates a new HTTP gateway.
func NewHttpGateway(method, urlTemplate, bodyTemplate string, headers map[string]string) (*httpGatewayImpl, error) {
	urlTmpl, err := NewTemplate("url", urlTemplate)
	if err != nil {
		return nil, fmt.Errorf("invalid URL template: %w", err)
	}

	bodyTmpl, err := NewTemplate("body", bodyTemplate)
	if err != nil {
		return nil, fmt.Errorf("invalid body template: %w", err)
	}

	return &httpGatewayImpl{
		method:       method,
		urlTemplate:  urlTmpl,
		bodyTemplate: bodyTmpl,
		headers:      headers,
		client:       newHttpClient(),
	}, nil
}

// NewHttpGatewayLegacy creates a new HTTP gateway using legacy token-based auth.
// DEPRECATED: Use NewHttpGateway with headers map instead.
func NewHttpGatewayLegacy(token, method, urlTemplate, bodyTemplate string) *httpGatewayImpl {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}

	gateway, err := NewHttpGateway(method, urlTemplate, bodyTemplate, headers)
	if err != nil {
		// Fallback to basic implementation if templates fail
		return &httpGatewayImpl{
			method:  method,
			headers: headers,
			client:  newHttpClient(),
		}
	}

	return gateway
}

func (hg *httpGatewayImpl) req(ctx context.Context, url string, body io.Reader, headers map[string]string) (Response, error) {
	hg.mu.RLock()
	method := hg.method
	hg.mu.RUnlock()

	switch method {
	case http.MethodGet:
		return hg.client.Get(ctx, url, headers)
	case http.MethodPost:
		return hg.client.Post(ctx, url, body, headers)
	case http.MethodPut:
		return hg.client.Put(ctx, url, body, headers)
	case http.MethodDelete:
		return hg.client.Delete(ctx, url, headers)
	case http.MethodPatch:
		return hg.client.Patch(ctx, url, body, headers)
	default:
		return Response{}, errors.New("method not supported: " + method)
	}
}

// Exec executes the request with the given variables to fill the body and url templates.
// It supports template rendering for both URL, body, and header values.
func (hg *httpGatewayImpl) Exec(ctx context.Context, variables map[string]string) (Response, error) {
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

// UpdateConfig updates the gateway configuration and templates at runtime (hot-reload).
func (hg *httpGatewayImpl) UpdateConfig(method, urlTemplate, bodyTemplate string, headers map[string]string) error {
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
