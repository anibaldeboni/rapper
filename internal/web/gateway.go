package web

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"text/template"
)

// httpGatewayImpl implements the processor.HttpGateway interface.
// This is an internal implementation - the interface is defined by the client (processor package).
type httpGatewayImpl struct {
	client          *httpClientImpl
	urlTemplate     *template.Template
	bodyTemplate    *template.Template
	headers         map[string]string // Flexible headers (Authorization, Cookie, etc)
	headerTemplates map[string]*template.Template
	fields          []string
	method          string
	mu              sync.RWMutex // Protects against concurrent access during hot-reload
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
	headerTemplates := compileHeaders(headers)

	return &httpGatewayImpl{
		method:          method,
		urlTemplate:     urlTmpl,
		bodyTemplate:    bodyTmpl,
		headers:         headers,
		headerTemplates: headerTemplates,
		fields:          templateFields(urlTmpl, bodyTmpl, headerTemplates),
		client:          newHttpClient(),
	}, nil
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
		tmpl, ok := hg.headerTemplates[key]
		if !ok {
			headers[key] = value
			continue
		}
		headers[key] = RenderTemplate(tmpl, variables).String()
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
	headerTemplates := compileHeaders(headers)

	// Update all fields atomically
	hg.method = method
	hg.urlTemplate = urlTmpl
	hg.bodyTemplate = bodyTmpl
	hg.headers = headers
	hg.headerTemplates = headerTemplates
	hg.fields = templateFields(urlTmpl, bodyTmpl, headerTemplates)

	return nil
}

func (hg *httpGatewayImpl) TemplateFields() []string {
	hg.mu.RLock()
	defer hg.mu.RUnlock()

	fields := make([]string, len(hg.fields))
	copy(fields, hg.fields)
	return fields
}

func compileHeaders(headers map[string]string) map[string]*template.Template {
	compiled := make(map[string]*template.Template, len(headers))
	for key, value := range headers {
		tmpl, err := NewTemplate("header-"+key, value)
		if err == nil {
			compiled[key] = tmpl
		}
	}
	return compiled
}

func templateFields(urlTmpl, bodyTmpl *template.Template, headers map[string]*template.Template) []string {
	fields := collectFields(urlTmpl)
	seen := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		seen[field] = struct{}{}
	}
	for _, field := range collectFields(bodyTmpl) {
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		fields = append(fields, field)
	}

	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		for _, field := range collectFields(headers[key]) {
			if _, ok := seen[field]; ok {
				continue
			}
			seen[field] = struct{}{}
			fields = append(fields, field)
		}
	}
	return fields
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
