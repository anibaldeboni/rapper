package web_test

import (
	"bytes"
	"rapper/files"
	"rapper/web"
	"rapper/web/mocks"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestExec(t *testing.T) {
	t.Run("should execute the request", func(t *testing.T) {
		httpClient := mocks.NewHttpClient(t)
		url := "api.site.domain/"
		body := `{ "key": "value" }`
		headers := map[string]string{"Authorization": "Bearer auth-token"}
		gateway := &web.HttpGatewayImpl{
			Token:  "auth-token",
			Method: "PUT",
			Client: httpClient,
			Templates: struct {
				Url  *template.Template
				Body *template.Template
			}{
				files.NewTemplate("url", url+"{{.id}}"),
				files.NewTemplate("body", `{ "key": "{{.value}}" }`),
			},
		}
		httpClient.On("Put", url+"1", bytes.NewBuffer([]byte(body)), headers).Return(web.Response{}, nil)

		variables := map[string]string{
			"id":    "1",
			"value": "value",
		}

		res, err := gateway.Exec(variables)

		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}
