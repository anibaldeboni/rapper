package cli_test

import (
	"testing"

	"github.com/anibaldeboni/rapper/cli"
	logMock "github.com/anibaldeboni/rapper/internal/log/mocks"
	processorMocks "github.com/anibaldeboni/rapper/internal/processor/mocks"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	logManagerMock := logMock.NewLogManager(t)
	processorMock := processorMocks.NewProcessor(t)
	t.Run("When the path contains CSV files", func(t *testing.T) {
		filePaths := []string{"../../tests/example.csv"}

		c := cli.New(filePaths, processorMock, logManagerMock)

		assert.NotNil(t, c)
	})
}

func TestUI(t *testing.T) {
	logManagerMock := logMock.NewLogManager(t)
	processorMock := processorMocks.NewProcessor(t)
	t.Run("Should quit when the user presses 'q'", func(t *testing.T) {
		filePaths := []string{"../../tests/example.csv"}

		m := cli.New(filePaths, processorMock, logManagerMock)

		tm := teatest.NewTestModel(
			t, m,
			teatest.WithInitialTermSize(300, 100),
		)

		tm.Send(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune("q"),
		})

		tm.WaitFinished(t)
	})
}
