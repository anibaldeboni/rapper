package cli_test

import (
	"rapper/cli"
	"rapper/files"
	uiMocks "rapper/ui/spinner/mocks"
	webMocks "rapper/web/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

var records = []files.CSVLine{
	{
		"id":   "1",
		"name": "John Doe",
	},
}
var csv = files.CSV{
	Name:  "test.csv",
	Lines: records,
}

func TestRun(t *testing.T) {
	t.Run("should call the spinner", func(t *testing.T) {
		spinner := uiMocks.NewSpinner(t)
		hg := webMocks.NewHttpGateway(t)

		spinner.On("Run").Return(nil, nil)

		err := cli.Run(csv, hg, spinner)
		assert.NoError(t, err)
	})
}
