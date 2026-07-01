package ui_test

import (
	"bytes"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anibaldeboni/rapper/internal/ui"
	mock_ui "github.com/anibaldeboni/rapper/internal/ui/mock"
	"github.com/anibaldeboni/rapper/internal/ui/ports"
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
		// Mock Processor.GetWorkerCount() which is called during settings view
		// initialization to seed the worker-count slider at the top of the view.
		processorMock.EXPECT().GetWorkerCount().Return(1).AnyTimes()
		// Mock Processor.GetMaxWorkers() which is called during settings view
		// initialization to bound the worker-count slider to runtime.NumCPU().
		processorMock.EXPECT().GetMaxWorkers().Return(1).AnyTimes()

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
	processorMock.EXPECT().GetMaxWorkers().Return(1).AnyTimes()
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

func TestProgramWithWindowSizeIntegration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logManagerMock := mock_ui.NewMockLogService(ctrl)
	processorMock := mock_ui.NewMockProcessorController(ctrl)
	configMgrMock := mock_ui.NewMockConfigManager(ctrl)

	filePaths := []string{"../../tests/example.csv"}

	logManagerMock.EXPECT().Get().Return([]string{}).AnyTimes()
	configMgrMock.EXPECT().Get().Return(nil).AnyTimes()
	configMgrMock.EXPECT().GetActiveProfile().Return("default").AnyTimes()
	configMgrMock.EXPECT().ListProfiles().Return([]string{"default"}).AnyTimes()
	processorMock.EXPECT().GetWorkerCount().Return(1).AnyTimes()
	processorMock.EXPECT().GetMaxWorkers().Return(1).AnyTimes()
	processorMock.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{}).AnyTimes()

	app := ui.NewApp(filePaths, processorMock, logManagerMock, configMgrMock)

	var in bytes.Buffer
	var out bytes.Buffer

	p := tea.NewProgram(
		app,
		tea.WithWindowSize(80, 24),
		tea.WithInput(&in),
		tea.WithOutput(&out),
	)

	// Use a forced shutdown after 200ms so the test does not depend
	// on the timing of `p.Send(tea.Quit)` (which is racy in
	// bubbletea v2 — the unbuffered p.msgs channel blocks the send
	// goroutine indefinitely if the program is not yet reading, and
	// the program's event loop can lose the race against the tick
	// chain's continuous TickMsg production). The 200ms delay gives
	// the program time to render at least once so the
	// `out.String()` assertion has content to check.
	go func() {
		time.Sleep(200 * time.Millisecond)
		p.Kill()
	}()

	model, err := p.Run()
	// p.Kill returns ErrProgramKilled; that's an acceptable outcome
	// for this test — the goal is to verify the program can be
	// constructed, started, and torn down, not to test the Quit
	// message path. The test was previously hanging on the Quit
	// send in bubbletea v2.
	if err != nil && err != tea.ErrProgramKilled {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.NotNil(t, model)
	assert.NotEmpty(t, out.String(), "program should render at least once before shutdown")
}
