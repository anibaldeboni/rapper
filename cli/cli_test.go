package cli_test

import (
	"bytes"
	"errors"
	"net/http"
	"rapper/cli"
	"rapper/files"
	"rapper/web"
	webMocks "rapper/web/mocks"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const path = "../tests"

func TestNew(t *testing.T) {
	t.Run("When the path contains CSV files", func(t *testing.T) {
		config := files.AppConfig{}

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

func TestUI(t *testing.T) {
	t.Run("Should quit when the user presses 'q'", func(t *testing.T) {
		config := files.AppConfig{}

		gatewayMock := webMocks.NewHttpGateway(t)

		m, err := cli.New(config, path, gatewayMock, "appName", "appVersion")
		assert.NoError(t, err)

		tm := teatest.NewTestModel(
			t, m,
			teatest.WithInitialTermSize(300, 100),
		)
		teatest.WaitFor(
			t, tm.Output(),
			func(bts []byte) bool {
				return bytes.Contains(bts, []byte("Choose a file to process"))
			},
			teatest.WithCheckInterval(time.Millisecond*100),
			teatest.WithDuration(time.Second*3),
		)

		tm.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("q"),
		})

		tm.WaitFinished(t)
	})

	t.Run("Should show 'done' when all requests complete", func(t *testing.T) {
		config := files.AppConfig{}

		gatewayMock := webMocks.NewHttpGateway(t)
		gatewayMock.On("Exec", mock.Anything).Return(web.Response{Status: http.StatusOK}, nil)

		m, err := cli.New(config, path, gatewayMock, "appName", "appVersion")
		assert.NoError(t, err)

		tm := teatest.NewTestModel(
			t, m,
			teatest.WithInitialTermSize(300, 100),
		)
		teatest.WaitFor(
			t, tm.Output(),
			func(bts []byte) bool {
				return bytes.Contains(bts, []byte("Choose a file to process"))
			},
			teatest.WithCheckInterval(time.Millisecond*100),
			teatest.WithDuration(time.Second*3),
		)

		tm.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("enter"),
		})

		teatest.WaitFor(
			t, tm.Output(),
			func(bts []byte) bool {
				return bytes.Contains(bts, []byte("done!"))
			},
			teatest.WithCheckInterval(time.Millisecond*100),
			teatest.WithDuration(time.Second*3),
		)
		err = tm.Quit()
		assert.NoError(t, err)

		tm.WaitFinished(t)
	})

	t.Run("Should show 'Request error' when the requests fail", func(t *testing.T) {
		config := files.AppConfig{}

		gatewayMock := webMocks.NewHttpGateway(t)
		gatewayMock.On("Exec", mock.Anything).Return(web.Response{}, errors.New("Request error"))

		m, err := cli.New(config, path, gatewayMock, "appName", "appVersion")
		assert.NoError(t, err)

		tm := teatest.NewTestModel(
			t, m,
			teatest.WithInitialTermSize(300, 100),
		)
		teatest.WaitFor(
			t, tm.Output(),
			func(bts []byte) bool {
				return bytes.Contains(bts, []byte("Choose a file to process"))
			},
			teatest.WithCheckInterval(time.Millisecond*100),
			teatest.WithDuration(time.Second*3),
		)

		tm.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("enter"),
		})

		teatest.WaitFor(
			t, tm.Output(),
			func(bts []byte) bool {
				return bytes.Contains(bts, []byte("Request error"))
			},
			teatest.WithCheckInterval(time.Millisecond*100),
			teatest.WithDuration(time.Second*3),
		)
		err = tm.Quit()
		assert.NoError(t, err)

		tm.WaitFinished(t)
	})
}
