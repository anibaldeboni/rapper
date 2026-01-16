package web_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/anibaldeboni/rapper/internal/web"
	mock_web "github.com/anibaldeboni/rapper/internal/web/mock"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func buildGateway(t *testing.T, client *mock_web.MockHttpClient, method string) *web.HttpGatewayImpl {
	t.Helper()
	
	gateway := &web.HttpGatewayImpl{
		Client: client,
	}
	
	// Initialize the gateway with proper configuration
	err := gateway.UpdateConfig(method, "api.site.domain/{{.id}}", `{ "key": "{{.value}}" }`, map[string]string{"Authorization": "Bearer auth-token"})
	if err != nil {
		t.Fatalf("failed to update gateway config: %v", err)
	}

	return gateway
}
func TestExec(t *testing.T) {
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
			method:  http.MethodPut,
			wantErr: true,
		},
		{
			name:    "should use the method post",
			method:  http.MethodPost,
			wantErr: false,
		},
		{
			name:    "should use the method put",
			method:  http.MethodPut,
			wantErr: false,
		},
		{
			name:    "should return error if unsupported method is used",
			method:  "UNSUPPORTED",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			httpClient := mock_web.NewMockHttpClient(ctrl)

		gateway := buildGateway(t, httpClient, tt.method)
		ctx := context.Background()

		var err error
		var res web.Response
		if tt.wantErr {
			err = errors.New("error")
		} else {
			res = successResponse
		}
		switch tt.method {
		case http.MethodPost:
			httpClient.EXPECT().Post(ctx, url+"1", bytes.NewBuffer([]byte(body)), headers).Return(res, err)
		case http.MethodPut:
			httpClient.EXPECT().Put(ctx, url+"1", bytes.NewBuffer([]byte(body)), headers).Return(res, err)
		}

		res, e := gateway.Exec(ctx, variables)

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
