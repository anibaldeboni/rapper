package ui_test

import (
	"testing"

	mock_log "github.com/anibaldeboni/rapper/internal/logs/mock"
	mock_processor "github.com/anibaldeboni/rapper/internal/processor/mock"
	"github.com/anibaldeboni/rapper/internal/ui"
	"go.uber.org/mock/gomock"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/assert"
)

func TestNewUI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logManagerMock := mock_log.NewMockLogger(ctrl)
	processorMock := mock_processor.NewMockProcessor(ctrl)

	t.Run("When the path contains CSV files", func(t *testing.T) {
		filePaths := []string{"../../tests/example.csv"}

		c := ui.New(filePaths, processorMock, logManagerMock)

		assert.NotNil(t, c)
	})
}

func TestUIQuit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logManagerMock := mock_log.NewMockLogger(ctrl)
	processorMock := mock_processor.NewMockProcessor(ctrl)

	t.Run("Should quit when the user presses 'q'", func(t *testing.T) {
		filePaths := []string{"../../tests/example.csv"}

		m := ui.New(filePaths, processorMock, logManagerMock)

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
