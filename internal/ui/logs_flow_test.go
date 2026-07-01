package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	mock_ui "github.com/anibaldeboni/rapper/internal/ui/mock"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/anibaldeboni/rapper/internal/ui/ports"
	"github.com/anibaldeboni/rapper/internal/ui/views"
	"go.uber.org/mock/gomock"
)

func TestLogsFlow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logManagerMock := mock_ui.NewMockLogService(ctrl)
	processorMock := mock_ui.NewMockProcessorController(ctrl)
	configMgrMock := mock_ui.NewMockConfigManager(ctrl)

	logManagerMock.EXPECT().Get().Return([]string{}).AnyTimes()
	configMgrMock.EXPECT().Get().Return(nil).AnyTimes()
	configMgrMock.EXPECT().GetActiveProfile().Return("default").AnyTimes()
	configMgrMock.EXPECT().ListProfiles().Return([]string{"default"}).AnyTimes()
	processorMock.EXPECT().GetWorkerCount().Return(1).AnyTimes()
	processorMock.EXPECT().GetMaxWorkers().Return(1).AnyTimes()
	processorMock.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{}).AnyTimes()

	app := NewApp([]string{"test.csv"}, processorMock, logManagerMock, configMgrMock)

	app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app.Update(tea.KeyPressMsg{Code: tea.KeyF2})

	for i := 0; i < 3; i++ {
		app.Update(msgs.TickMsg(time.Now()))
	}

	logManagerMock.EXPECT().Get().Return([]string{"Test log message"}).AnyTimes()

	for i := 0; i < 3; i++ {
		app.Update(msgs.TickMsg(time.Now()))
	}

	logsView := app.views[ViewLogs].(views.LogsView)
	fmt.Println("Viewport width:", logsView.ViewportWidth())
	fmt.Println("Viewport height:", logsView.ViewportHeight())
	fmt.Println("Viewport content set:", logsView.ViewportContent())

	lv := logsView.View()
	fmt.Println("=== LogsView Content ===")
	fmt.Println(lv.Content)
	fmt.Println("=== End Content ===")
	fmt.Println("Has 'Test log message':", strings.Contains(lv.Content, "Test log message"))
}
