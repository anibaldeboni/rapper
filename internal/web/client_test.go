package web_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anibaldeboni/rapper/internal/web"

	"github.com/stretchr/testify/assert"
)

const noHeader = "/no-header"
const withHeader = "/with-header"
const invalidRoute = "invalid"

func TestPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case noHeader:
			if r.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
				t.Errorf("Expected Content-Type: application/json header, got: %s", r.Header.Get("Content-Type"))
			}
		case withHeader:
			defaultHeader := r.Header.Get("Content-Type")
			additionalHeader := r.Header.Get("X-Test")
			if defaultHeader == "" || additionalHeader == "" {
				t.Errorf("Expected headers: Content-Type and X-test, got: %s %s", defaultHeader, additionalHeader)
			}
		default:
			t.Errorf("Invalid route: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"key":"value"}`))
		if err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	client := web.NewHttpClient()
	body := bytes.NewBuffer([]byte(`{ "key": "value" }`))

	tests := []struct {
		name    string
		route   string
		headers map[string]string
		wantErr bool
	}{
		{
			name:  "When no additional header is informed",
			route: server.URL + noHeader,
		},
		{
			name:    "When additional header is informed",
			route:   server.URL + withHeader,
			headers: map[string]string{"X-Test": "test"},
		},
		{
			name:    "When additional header is informed",
			route:   "http://invalid-domain/api",
			headers: map[string]string{"X-Test": "test"},
			wantErr: true,
		},
		{
			name:    "When URL is malformed",
			route:   server.URL + invalidRoute,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.Put(context.Background(), tt.route, body, tt.headers)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantErr, res.StatusCode != http.StatusOK)
		})
	}
}

func TestPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case noHeader:
			if r.Header.Get("Content-Type") != "application/json; charset=UTF-8" {
				t.Errorf("Expected Content-Type: application/json header, got: %s", r.Header.Get("Content-Type"))
			}
		case withHeader:
			defaultHeader := r.Header.Get("Content-Type")
			additionalHeader := r.Header.Get("X-Test")
			if defaultHeader == "" || additionalHeader == "" {
				t.Errorf("Expected headers: Content-Type and X-test, got: %s %s", defaultHeader, additionalHeader)
			}
		default:
			t.Errorf("Invalid route: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"key":"value"}`))
		if err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	client := web.NewHttpClient()
	body := bytes.NewBuffer([]byte(`{ "key": "value" }`))

	tests := []struct {
		name    string
		route   string
		headers map[string]string
		wantErr bool
	}{
		{
			name:  "When no additional header is informed",
			route: server.URL + noHeader,
		},
		{
			name:    "When additional header is informed",
			route:   server.URL + withHeader,
			headers: map[string]string{"X-Test": "test"},
		},
		{
			name:    "When additional header is informed",
			route:   "http://invalid-domain/api",
			headers: map[string]string{"X-Test": "test"},
			wantErr: true,
		},
		{
			name:    "When URL is malformed",
			route:   server.URL + invalidRoute,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.Post(context.Background(), tt.route, body, tt.headers)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantErr, res.StatusCode != http.StatusOK)
		})
	}
}

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case noHeader:
			var headers []string
			for name := range r.Header {
				if name != "User-Agent" && name != "Accept-Encoding" {
					headers = append(headers, name)
				}
			}
			if len(headers) != 0 {
				t.Errorf("Expected no headers, got: %s", headers)
			}
		case withHeader:
			if r.Header.Get("X-Test") == "" {
				t.Errorf("Expected headers: X-test to empty, got: %s", r.Header.Get("X-Test"))
			}
		default:
			t.Errorf("Invalid route: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"key":"value"}`))
		if err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	client := web.NewHttpClient()

	tests := []struct {
		name    string
		route   string
		headers map[string]string
		wantErr bool
	}{
		{
			name:  "When no additional header is informed",
			route: server.URL + noHeader,
		},
		{
			name:    "When additional header is informed",
			route:   server.URL + withHeader,
			headers: map[string]string{"X-Test": "test"},
		},
		{
			name:    "When additional header is informed",
			route:   "http://invalid-domain/api",
			headers: map[string]string{"X-Test": "test"},
			wantErr: true,
		},
		{
			name:    "When URL is malformed",
			route:   server.URL + invalidRoute,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := client.Get(context.Background(), tt.route, tt.headers)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantErr, res.StatusCode != http.StatusOK)
		})
	}
}
