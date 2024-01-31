package cli_test

import (
	"errors"
	"net/http"
	"rapper/cli"
	"rapper/files"
	uiMocks "rapper/ui/spinner/mocks"
	"rapper/web"
	webMocks "rapper/web/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	t.Run("When the request succeed", func(t *testing.T) {
		spinner := uiMocks.NewSpinner(t)
		hg := webMocks.NewHttpGateway(t)

		spinner.On("Run").Return(nil, nil)
		spinner.On("Update", mock.Anything).Return(nil)
		hg.On("Exec", mock.Anything).Return(web.Response{
			Status:  200,
			Body:    []byte(""),
			Headers: http.Header{},
		}, nil)

		err := cli.Run(csv, hg, spinner)
		assert.NoError(t, err)
	})
}

func TestRunWithInvalidRequestStatus(t *testing.T) {
	t.Run("When the request is status is not 200", func(t *testing.T) {
		spinner := uiMocks.NewSpinner(t)
		hg := webMocks.NewHttpGateway(t)

		spinner.On("Run").Return(nil, nil)
		spinner.On("Update", mock.Anything).Return(nil)
		hg.On("Exec", mock.Anything).Return(web.Response{
			Status:  401,
			Body:    []byte(""),
			Headers: http.Header{},
		}, nil)

		err := cli.Run(csv, hg, spinner)
		assert.NoError(t, err)
	})
}

func TestRunWithInvalidRequestError(t *testing.T) {
	t.Run("When the request connection fails", func(t *testing.T) {
		spinner := uiMocks.NewSpinner(t)
		hg := webMocks.NewHttpGateway(t)

		spinner.On("Run").Return(nil, nil)
		spinner.On("Update", mock.Anything).Return(nil)
		hg.On("Exec", mock.Anything).Return(web.Response{}, errors.New("request-error"))

		err := cli.Run(csv, hg, spinner)
		assert.NoError(t, err)
	})
}

func TestRunWithUIError(t *testing.T) {
	t.Run("When the ui brakes", func(t *testing.T) {
		spinner := uiMocks.NewSpinner(t)
		hg := webMocks.NewHttpGateway(t)

		spinner.On("Run").Return(nil, errors.New("ui-error"))

		err := cli.Run(csv, hg, spinner)
		spinner.AssertNotCalled(t, "Update", mock.Anything)
		hg.AssertNotCalled(t, "Exec", mock.Anything)
		assert.Error(t, err)
	})
}
