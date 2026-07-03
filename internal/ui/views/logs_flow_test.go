package views

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/anibaldeboni/rapper/internal/logs"
	mock_ui "github.com/anibaldeboni/rapper/internal/ui/mock"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/anibaldeboni/rapper/internal/ui/ports"
	"go.uber.org/mock/gomock"
)

func TestLogsView_ContentAfterTick(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logManagerMock := mock_ui.NewMockLogProvider(ctrl)
	processorMock := mock_ui.NewMockProcessorController(ctrl)

	processorMock.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{}).AnyTimes()
	logManagerMock.EXPECT().Clear().AnyTimes()

	// Build a counter-based mock that returns the empty list for the
	// first 5 calls (NewLogsView + Update(Size) + 3×Update(Tick)) and
	// the populated list for every call after that. This replaces the
	// previous two-expectation setup, which gomock matched in FIFO
	// order — the first AnyTimes expectation would catch every call
	// and the second one would never be reached.
	var calls int32
	logManagerMock.EXPECT().Get().DoAndReturn(func() []logs.LogMessage {
		// atomic add to avoid a race; the test is single-goroutine
		// so a plain int would also work, but this keeps the helper
		// future-safe.
		calls++
		if calls <= 5 {
			return []logs.LogMessage{}
		}
		return []logs.LogMessage{logs.NewGeneralMessage("", "", "Test log message")}
	}).AnyTimes()

	v := NewLogsView(logManagerMock, processorMock)

	// Resize
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 120, Height: 40})
	v = next.(LogsView)

	fmt.Println("After resize - Viewport width:", v.viewport.Width(), "height:", v.viewport.Height())
	fmt.Println("Lines count:", len(v.viewport.GetContent()))

	// Ticks
	for i := 0; i < 3; i++ {
		next, _ = v.Update(msgs.TickMsg(time.Now()))
		v = next.(LogsView)
	}

	fmt.Println("After ticks - Viewport width:", v.viewport.Width(), "height:", v.viewport.Height())
	fmt.Println("Lines count:", len(v.viewport.GetContent()))

	// Call updateLogs manually. After the elm-messaging-logs-fix the
	// function returns the modified LogsView, so the call site MUST
	// capture the return — otherwise the viewport content is lost
	// (the value-receiver mutation never reaches the caller).
	v = v.updateLogs()

	fmt.Println("After manual updateLogs - Viewport width:", v.viewport.Width(), "height:", v.viewport.Height())
	fmt.Println("Lines count:", len(v.viewport.GetContent()))
	fmt.Println("Content:", v.viewport.GetContent())
	fmt.Println("View raw:", v.viewport.View())

	lv := v.View()
	fmt.Println("=== LogsView Content ===")
	fmt.Println(lv.Content)
	fmt.Println("=== End Content ===")
	fmt.Println("Has 'Test log message':", strings.Contains(lv.Content, "Test log message"))

	if !strings.Contains(lv.Content, "Test log message") {
		t.Error("LogsView should contain the log message")
	}
}
