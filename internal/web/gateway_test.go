package web_test

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"testing"
	"text/template"

	"github.com/anibaldeboni/rapper/internal/files"
	"github.com/anibaldeboni/rapper/internal/web"
	"github.com/anibaldeboni/rapper/internal/web/mocks"

	"github.com/stretchr/testify/assert"
)

func buildGateway(t *testing.T, method string, client *mocks.HttpClient) *web.HttpGatewayImpl {
	t.Helper()
	gateway := &web.HttpGatewayImpl{
		Token:  "auth-token",
		Method: method,
		Client: client,
		Templates: struct {
			Url  *template.Template
			Body *template.Template
		}{
			files.NewTemplate("url", "api.site.domain/{{.id}}"),
			files.NewTemplate("body", `{ "key": "{{.value}}" }`),
		},
	}

	return gateway
}
func TestExec(t *testing.T) {
	httpClient := mocks.NewHttpClient(t)
	url := "api.site.domain/"
	body := `{ "key": "value" }`
	headers := map[string]string{"Authorization": "Bearer auth-token"}
	variables := map[string]string{
		"id":    "1",
		"value": "value",
	}
	successResponse := web.Response{
		StatusCode: 200,
		Body:       []byte(body),
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
	}

	tests := []struct {
		name    string
		method  string
		wantErr bool
	}{
		{
			name:    "should return error if the request fails",
			method:  "Put",
			wantErr: true,
		},
		{
			name:    "should use the method post",
			method:  "Post",
			wantErr: false,
		},
		{
			name:    "should use the method put",
			method:  "Put",
			wantErr: false,
		},
		{
			name:    "should return error if unsupported method is used",
			method:  "Get",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gateway := buildGateway(t, strings.ToUpper(tt.method), httpClient)

			var err error
			var res web.Response
			if tt.wantErr {
				err = errors.New("error")
			} else {
				res = successResponse
			}

			mockCall := httpClient.On(tt.method, url+"1", bytes.NewBuffer([]byte(body)), headers).Return(res, err)

			res, e := gateway.Exec(variables)
			if tt.method != "Get" {
				httpClient.AssertExpectations(t)
			}
			mockCall.Unset()

			if tt.wantErr {
				assert.Error(t, e)
				assert.Zero(t, res)
			} else {
				assert.NoError(t, e)
				assert.NotZero(t, res)
			}

		})
	}
}
