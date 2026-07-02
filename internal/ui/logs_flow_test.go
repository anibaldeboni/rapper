package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/anibaldeboni/rapper/internal/ui/views"
)

func TestLogsFlow(t *testing.T) {
	app, logManagerMock, _, _ := newTestApp(t, "test.csv")

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
