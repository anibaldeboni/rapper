package cli_test

import (
	"rapper/cli"
	"rapper/files"
	webMocks "rapper/web/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("When the path contains CSV files", func(t *testing.T) {
		config := files.AppConfig{}
		path := "../tests"

		c, err := cli.New(config, path, webMocks.NewHttpGateway(t), "appName", "appVersion")

		assert.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("When the path does not contain CSV files", func(t *testing.T) {
		config := files.AppConfig{}
		path := "../tests/empty"

		_, err := cli.New(config, path, webMocks.NewHttpGateway(t), "appName", "appVersion")

		assert.Error(t, err)
	})
}
