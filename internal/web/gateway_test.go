package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExec(t *testing.T) {
	t.Run("should execute POST requests", func(t *testing.T) {
		var seenMethod string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seenMethod = r.Method
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}))
		defer server.Close()

		gateway, err := NewHttpGateway(http.MethodPost, server.URL+"/{{.id}}", `{ "key": "{{.value}}" }`, map[string]string{"Authorization": "Bearer auth-token"})
		if !assert.NoError(t, err) {
			return
		}

		res, err := gateway.Exec(context.Background(), map[string]string{"id": "1", "value": "value"})
		assert.NoError(t, err)
		assert.Equal(t, http.MethodPost, seenMethod)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("should execute PUT requests", func(t *testing.T) {
		var seenMethod string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seenMethod = r.Method
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}))
		defer server.Close()

		gateway, err := NewHttpGateway(http.MethodPut, server.URL+"/{{.id}}", `{ "key": "{{.value}}" }`, map[string]string{"Authorization": "Bearer auth-token"})
		if !assert.NoError(t, err) {
			return
		}

		res, err := gateway.Exec(context.Background(), map[string]string{"id": "1", "value": "value"})
		assert.NoError(t, err)
		assert.Equal(t, http.MethodPut, seenMethod)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("should return error for unsupported method", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}))
		defer server.Close()

		gateway, err := NewHttpGateway("UNSUPPORTED", server.URL+"/{{.id}}", `{ "key": "{{.value}}" }`, map[string]string{"Authorization": "Bearer auth-token"})
		if !assert.NoError(t, err) {
			return
		}

		res, err := gateway.Exec(context.Background(), map[string]string{"id": "1", "value": "value"})
		assert.Error(t, err)
		assert.Zero(t, res)
	})
}
