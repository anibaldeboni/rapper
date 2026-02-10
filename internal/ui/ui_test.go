package ui_test

import (
	"testing"

	"github.com/anibaldeboni/rapper/internal/ui"
	mock_ui "github.com/anibaldeboni/rapper/internal/ui/mock"
	"github.com/anibaldeboni/rapper/internal/ui/ports"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func TestNewUI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logManagerMock := mock_ui.NewMockLogService(ctrl)
	processorMock := mock_ui.NewMockProcessorController(ctrl)
	configMgrMock := mock_ui.NewMockConfigManager(ctrl)

	t.Run("When the path contains CSV files", func(t *testing.T) {
		filePaths := []string{"../../tests/example.csv"}

		// Mock Logger.Get() which is called during UI initialization in updateLogs()
		logManagerMock.EXPECT().Get().Return([]string{}).AnyTimes()
		// Mock ConfigManager.Get() which is called during settings view initialization
		configMgrMock.EXPECT().Get().Return(nil).AnyTimes()
		// Mock Processor.GetWorkerCount() which is called during workers view initialization
		processorMock.EXPECT().GetWorkerCount().Return(1).AnyTimes()

		c := ui.NewApp(filePaths, processorMock, logManagerMock, configMgrMock)

		assert.NotNil(t, c)
	})
}
func TestSmallWindowHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	logManagerMock := mock_ui.NewMockLogService(ctrl)
	processorMock := mock_ui.NewMockProcessorController(ctrl)
	configMgrMock := mock_ui.NewMockConfigManager(ctrl)

	filePaths := []string{"../../tests/example.csv"}

	// Setup mocks
	logManagerMock.EXPECT().Get().Return([]string{}).AnyTimes()
	configMgrMock.EXPECT().Get().Return(nil).AnyTimes()
	configMgrMock.EXPECT().GetActiveProfile().Return("default").AnyTimes()
	configMgrMock.EXPECT().ListProfiles().Return([]string{"default"}).AnyTimes()
	processorMock.EXPECT().GetWorkerCount().Return(1).AnyTimes()
	processorMock.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{}).AnyTimes()

	app := ui.NewApp(filePaths, processorMock, logManagerMock, configMgrMock)

	testCases := []struct {
		name   string
		width  int
		height int
	}{
		{"Very small window", 40, 10},
		{"Small height", 80, 15},
		{"Small width", 50, 24},
		{"Minimum viable", 30, 8},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Send window resize message
			msg := tea.WindowSizeMsg{Width: tc.width, Height: tc.height}

			// Update should not panic
			assert.NotPanics(t, func() {
				app.Update(msg)
			})

			// View should render without panic
			assert.NotPanics(t, func() {
				view := app.View()
				assert.NotEmpty(t, view)
			})
		})
	}
}
