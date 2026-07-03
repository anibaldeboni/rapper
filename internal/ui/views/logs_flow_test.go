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
	// first call and a populated list thereafter. The previous
	// two-expectation setup would gomock-match in FIFO order — the
	// first AnyTimes would catch every call and the second one would
	// never be reached.
	var calls int32
	logManagerMock.EXPECT().Get().DoAndReturn(func() []logs.LogMessage {
		calls++
		if calls == 1 {
			return []logs.LogMessage{}
		}
		return []logs.LogMessage{logs.NewGeneralMessage("", "", "Test log message")}
	}).AnyTimes()

	v := NewLogsView(logManagerMock, processorMock)

	// Resize
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 120, Height: 40})
	v = next.(LogsView)

	fmt.Println("After resize - list width:", v.list.Width(), "height:", v.list.Height())
	fmt.Println("Items count:", v.list.Len())

	// Tick
	next, _ = v.Update(msgs.MetricsTickMsg(time.Now()))
	v = next.(LogsView)

	fmt.Println("After tick - list width:", v.list.Width(), "height:", v.list.Height())
	fmt.Println("Items count:", v.list.Len())
	fmt.Println("Content:", v.View().Content)

	lv := v.View()
	fmt.Println("=== LogsView Content ===")
	fmt.Println(lv.Content)
	fmt.Println("=== End Content ===")
	fmt.Println("Has 'Test log message':", strings.Contains(lv.Content, "Test log message"))

	if !strings.Contains(lv.Content, "Test log message") {
		t.Error("LogsView should contain the log message")
	}
}
