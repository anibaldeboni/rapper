package views

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/ui/components"
	mock_ui "github.com/anibaldeboni/rapper/internal/ui/mock"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/anibaldeboni/rapper/internal/ui/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// longestVisibleLine returns the visible width of the widest line in s.
// Used by the metrics-column-width tests to assert that the column
// allocated to the metrics panel can hold its widest row without
// clipping.
func longestVisibleLine(s string) int {
	max := 0
	for _, line := range strings.Split(s, "\n") {
		w := lipgloss.Width(line)
		if w > max {
			max = w
		}
	}
	return max
}

// newTestLogsView builds a LogsView with gomock-backed LogProvider
// and ProcessorController. The default GetMetrics / Get returns
// (ProcessorMetrics{}, []LogMessage{}); tests that need richer data
// register their own expectations.
//
//nolint:unparam // logger is consumed by callers that need to set up richer expectations
func newTestLogsView(t *testing.T) (LogsView, *mock_ui.MockLogProvider, *mock_ui.MockProcessorController) {
	t.Helper()
	ctrl := gomock.NewController(t)
	proc := mock_ui.NewMockProcessorController(ctrl)
	logger := mock_ui.NewMockLogProvider(ctrl)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{}).AnyTimes()
	logger.EXPECT().Get().Return([]logs.LogMessage{}).AnyTimes()
	logger.EXPECT().Clear().AnyTimes()
	return NewLogsView(logger, proc), logger, proc
}

func TestLogsView_Resize_AllocatesEnoughWidthForMetricsPanel(t *testing.T) {
	v, _, proc := newTestLogsView(t)
	// Realistic snapshot: large request counts push the throughput and
	// totals to their widest possible rendering. Re-register an
	// expectation with a richer value for this test.
	gomock.NewController(t)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{
		TotalRequests:   12345,
		SuccessRequests: 12300,
		ErrorRequests:   45,
		LinesProcessed:  12345,
		ActiveWorkers:   8,
		RequestsPerSec:  123.45,
		IsProcessing:    true,
	}).AnyTimes()

	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 120, Height: 40})
	v = next.(LogsView)

	// The metrics panel must be at least as wide as its longest rendered
	// row, otherwise lipgloss clips the right edge of every row that
	// exceeds the column width.
	requiredWidth := longestVisibleLine(v.metrics.View().Content)
	assert.LessOrEqualf(t, requiredWidth, v.rightCol,
		"metrics column width (%d) must fit the longest metrics row (width %d); increase metricsDefaultWidth",
		v.rightCol, requiredWidth)
}

func TestLogsView_Resize_DefaultMetricsWidthExceedsSmallestLabels(t *testing.T) {
	v, _, _ := newTestLogsView(t)

	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 120, Height: 40})
	v = next.(LogsView)

	assert.Greater(t, v.rightCol, 20, "metrics column must accommodate the 20-char label column plus a value")
}

// TestLogsView_Resize_AccountsForMarginLeft is the regression test for
// BUG-001: LogsView's View() applies MarginLeft(2) to the rendered body,
// but the partition math must respect it. With Resize(80, 20) and a
// fixed right pane of 24, the list width must be 80 - 24 - 2 = 54.
func TestLogsView_Resize_AccountsForMarginLeft(t *testing.T) {
	v, _, _ := newTestLogsView(t)

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{name: "narrow 40-col", width: 40, height: 20},
		{name: "common 80-col", width: 80, height: 20},
		{name: "wide 120-col", width: 120, height: 20},
		{name: "extra-wide 200-col", width: 200, height: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next, _ := v.Update(msgs.ViewportSizeMsg{Width: tt.width, Height: tt.height})
			v = next.(LogsView)

			listWidth, listHeight := v.listWidthHeight()
			// The View() body joins the list and the metrics panel and
			// then applies MarginLeft(2), so the invariant for an
			// exact fit is:
			//   listWidth + rightCol + logsMarginLeft == width
			assert.Equal(t, tt.width, listWidth+v.rightCol+logsMarginLeft,
				"logs body (left + right + marginLeft) must equal the assigned width")
			assert.LessOrEqual(t, listHeight, tt.height, "list height must not exceed assigned height")
		})
	}
}

// TestLogsView_Update_ViewportSizeMsg_PartitionInvariant — Phase 2
// version of the partition regression test. The invariant holds
// (listWidth + rightCol + logsMarginLeft == assigned width).
func TestLogsView_Update_ViewportSizeMsg_PartitionInvariant(t *testing.T) {
	v, _, _ := newTestLogsView(t)

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{name: "narrow 40-col", width: 40, height: 20},
		{name: "common 80-col", width: 80, height: 20},
		{name: "wide 120-col", width: 120, height: 20},
		{name: "extra-wide 200-col", width: 200, height: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next, _ := v.Update(msgs.ViewportSizeMsg{Width: tt.width, Height: tt.height})
			v = next.(LogsView)
			listWidth, _ := v.listWidthHeight()
			assert.Equal(t, tt.width, listWidth+v.rightCol+logsMarginLeft,
				"partition must be exact: left + right + marginLeft == assigned width")
		})
	}
}

// TestLogsView_Update_MetricsVisibilityMsg_StartsTick — when
// MetricsVisibilityMsg{Visible: true} arrives, the returned cmd must
// be non-nil (a tea.Tick that produces MetricsTickMsg).
func TestLogsView_Update_MetricsVisibilityMsg_StartsTick(t *testing.T) {
	v, _, _ := newTestLogsView(t)

	next, cmd := v.Update(msgs.MetricsVisibilityMsg{Visible: true})
	v = next.(LogsView)

	if cmd == nil {
		t.Fatal("MetricsVisibilityMsg{Visible: true} must return a non-nil tick cmd")
	}
	if !v.metrics.Visible {
		t.Errorf("MetricsPanel.Visible must be true after starting tick; got false")
	}
}

// TestLogsView_Update_MetricsVisibilityMsg_StopsTick — when
// MetricsVisibilityMsg{Visible: false} arrives, the returned cmd must
// be nil and the metrics panel Visible flag must be false.
func TestLogsView_Update_MetricsVisibilityMsg_StopsTick(t *testing.T) {
	v, _, _ := newTestLogsView(t)

	next, _ := v.Update(msgs.MetricsVisibilityMsg{Visible: true})
	v = next.(LogsView)
	next, cmd := v.Update(msgs.MetricsVisibilityMsg{Visible: false})
	v = next.(LogsView)

	if cmd != nil {
		t.Errorf("MetricsVisibilityMsg{Visible: false} must return nil cmd; got %T", cmd)
	}
	if v.metrics.Visible {
		t.Errorf("MetricsPanel.Visible must be false after stopping tick; got true")
	}
}

// TestLogsView_Init_ReturnsNil — Init must return nil per R-6.
func TestLogsView_Init_ReturnsNil(t *testing.T) {
	v, _, _ := newTestLogsView(t)
	if cmd := v.Init(); cmd != nil {
		t.Fatalf("Init must return nil; got %T", cmd)
	}
}

// TestLogsView_MetricsTick_AppendsNewLogs — the MetricsTickMsg
// handler must append any new log messages to the embedded list. The
// LogProvider is configured to return an empty list, then a one-row
// list, then a two-row list. Each MetricsTickMsg should grow the
// embedded list by exactly the right amount.
//
// The mock uses DoAndReturn with a counter so the test can assert
// the grow-on-tick behaviour deterministically without AnyTimes
// ambiguity.
func TestLogsView_MetricsTick_AppendsNewLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	proc := mock_ui.NewMockProcessorController(ctrl)
	logger := mock_ui.NewMockLogProvider(ctrl)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{IsProcessing: true}).AnyTimes()

	var calls int32
	logger.EXPECT().Get().DoAndReturn(func() []logs.LogMessage {
		calls++
		switch calls {
		case 1:
			return []logs.LogMessage{logs.NewGeneralMessage("💃", "Processing", "starting")}
		default:
			return []logs.LogMessage{
				logs.NewGeneralMessage("💃", "Processing", "starting"),
				logs.NewGeneralMessage("ℹ️", "Request", "second"),
			}
		}
	}).AnyTimes()
	logger.EXPECT().Clear().AnyTimes()

	v := NewLogsView(logger, proc)
	// Resize so the view has a non-zero size.
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})
	v = next.(LogsView)

	// Trigger the first tick — the list should land on 1 item.
	next, _ = v.Update(msgs.MetricsTickMsg{})
	v = next.(LogsView)
	oneItemView := v.View().Content
	require.Contains(t, oneItemView, "starting", "first tick must render the first message")

	// Trigger a second tick — the list should grow to 2 items.
	next, _ = v.Update(msgs.MetricsTickMsg{})
	v = next.(LogsView)
	twoItemView := v.View().Content
	require.Contains(t, twoItemView, "second", "second tick must render the second message")
}

// TestLogsView_ProcessingStartedMsg_ResetsList — the
// ProcessingStartedMsg handler must call DetailedList.Reset() AND
// logger.Clear() so each run starts from a clean slate. The list
// alone is not enough — the in-memory buffer must be cleared too,
// otherwise the next MetricsTickMsg would re-append the old
// messages.
func TestLogsView_ProcessingStartedMsg_ResetsList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	proc := mock_ui.NewMockProcessorController(ctrl)
	logger := mock_ui.NewMockLogProvider(ctrl)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{IsProcessing: true}).AnyTimes()
	logger.EXPECT().Clear().AnyTimes()

	// The mock returns an ever-growing list of "first run" messages
	// so every Get() that fires before the ProcessingStartedMsg lands
	// in the list. After the reset, the same mock returns a single
	// "fresh run" message.
	var calls int32
	logger.EXPECT().Get().DoAndReturn(func() []logs.LogMessage {
		calls++
		if calls <= 3 {
			// NewLogsView + 2 ticks each return progressively more
			// messages so the list grows visibly. The exact
			// count is not asserted (we only check the reset
			// behaviour).
			return make([]logs.LogMessage, calls)
		}
		return []logs.LogMessage{logs.NewGeneralMessage("💃", "Processing", "fresh run")}
	}).AnyTimes()

	v := NewLogsView(logger, proc)
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})
	v = next.(LogsView)

	// Two ticks — list grows.
	next, _ = v.Update(msgs.MetricsTickMsg{})
	v = next.(LogsView)
	next, _ = v.Update(msgs.MetricsTickMsg{})
	v = next.(LogsView)
	require.Greater(t, v.list.Len(), 0, "precondition: list has at least one item after the first two ticks")

	// Send ProcessingStartedMsg — list must reset and logger must clear.
	next, _ = v.Update(msgs.ProcessingStartedMsg{FilePath: "x.csv"})
	v = next.(LogsView)
	assert.Equal(t, 0, v.list.Len(), "list must be empty immediately after ProcessingStartedMsg")

	// A tick is needed for the new message to land in the list
	// (the Reset clears, then the next refresh appends).
	next, _ = v.Update(msgs.MetricsTickMsg{})
	v = next.(LogsView)

	assert.Equal(t, 1, v.list.Len(), "list must contain only the new run's first message after reset + tick")
	out := v.View().Content
	assert.Contains(t, out, "fresh run", "list must show the new run's first message")
}

// TestMetricsPanel_Render_IsDeterministicForGivenMetrics is a sanity
// test that the component itself emits a stable width. If a future
// change ever makes the panel wider, the LogsView column-width test
// above will fail and force a decision rather than letting the
// regression reach the user.
func TestMetricsPanel_Render_IsDeterministicForGivenMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	proc := mock_ui.NewMockProcessorController(ctrl)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{
		TotalRequests:   1,
		SuccessRequests: 1,
		ErrorRequests:   0,
		LinesProcessed:  1,
		ActiveWorkers:   1,
		RequestsPerSec:  0.5,
		IsProcessing:    false,
	}).AnyTimes()

	p := components.NewMetricsPanel(proc)
	first := p.View()
	second := p.View()
	assert.Equal(t, first, second, "metrics panel must be deterministic for the same metrics")
}
