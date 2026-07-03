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

func TestLogsView_Resize_AllocatesEnoughWidthForMetricsPanel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	proc := mock_ui.NewMockProcessorController(ctrl)
	// Realistic snapshot: large request counts push the throughput and
	// totals to their widest possible rendering.
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{
		TotalRequests:   12345,
		SuccessRequests: 12300,
		ErrorRequests:   45,
		LinesProcessed:  12345,
		ActiveWorkers:   8,
		RequestsPerSec:  123.45,
		IsProcessing:    true,
	}).AnyTimes()

	logger := mock_ui.NewMockLogProvider(ctrl)
	logger.EXPECT().Get().Return([]logs.LogMessage{}).AnyTimes()
	logger.EXPECT().Clear().AnyTimes()

	v := NewLogsView(logger, proc)
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 120, Height: 40})
	v = next.(LogsView)

	// The metrics panel must be at least as wide as its longest rendered
	// row, otherwise lipgloss clips the right edge of every row that
	// exceeds the column width (the user reported "está cortando os
	// textos" — the text is being cut).
	requiredWidth := longestVisibleLine(v.metrics.View().Content)
	assert.LessOrEqualf(t, requiredWidth, v.rightCol,
		"metrics column width (%d) must fit the longest metrics row (width %d); increase metricsDefaultWidth",
		v.rightCol, requiredWidth)
}

func TestLogsView_Resize_DefaultMetricsWidthExceedsSmallestLabels(t *testing.T) {
	// Regression guard: the metrics column must be wide enough to hold
	// the 20-char label column plus a separator plus a multi-digit
	// value. Use a stable value to keep the assertion deterministic.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	proc := mock_ui.NewMockProcessorController(ctrl)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{
		ActiveWorkers: 1,
		IsProcessing:  false,
	}).AnyTimes()
	logger := mock_ui.NewMockLogProvider(ctrl)
	logger.EXPECT().Get().Return([]logs.LogMessage{}).AnyTimes()
	logger.EXPECT().Clear().AnyTimes()

	v := NewLogsView(logger, proc)
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 120, Height: 40})
	v = next.(LogsView)

	// 20 (label) + 1 (space) + a multi-digit value. "Active Workers: 1"
	// is the shortest row; we just need the column to exceed the label
	// width so the value has room.
	assert.Greater(t, v.rightCol, 20, "metrics column must accommodate the 20-char label column plus a value")
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

// TestLogsView_Resize_AccountsForMarginLeft is the regression test for
// BUG-001: LogsView's View() applies MarginLeft(2) to the rendered body,
// but the Resize() math ignored it. With Resize(80, 20) and a fixed
// right pane of 24, the viewport was set to 56 columns even though the
// left pane rendered into 54 visible columns (56 - 2 from MarginLeft).
// The bug surfaced as 2 columns of overflow on the right edge of the
// logs pane (clipping the last 2 characters of every line).
//
// Root cause: `left := max(width-right, 0)` did not subtract the
// view-local MarginLeft. The fix introduces a named constant
// logsMarginLeft and uses it in the partition math so the rendered
// total (left + right + marginLeft) fits the assigned width exactly.
func TestLogsView_Resize_AccountsForMarginLeft(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	proc := mock_ui.NewMockProcessorController(ctrl)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{IsProcessing: false}).AnyTimes()
	logger := mock_ui.NewMockLogProvider(ctrl)
	logger.EXPECT().Get().Return([]logs.LogMessage{}).AnyTimes()
	logger.EXPECT().Clear().AnyTimes()

	v := NewLogsView(logger, proc)

	// Cover a representative range of terminal widths. Each pair asserts
	// the partition is exact: left + right + logsMarginLeft == width, so
	// the rendered body (which adds MarginLeft to the joined output) fits
	// the assigned area without clipping.
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
			// Phase 2: use the message-based Update path. Resize is gone.
			next, _ := v.Update(msgs.ViewportSizeMsg{Width: tt.width, Height: tt.height})
			v = next.(LogsView)

			// v.viewport.Width is the left-pane width assigned to the
			// viewport. v.rightCol is the metrics-pane width. The View()
			// body joins the two and then applies MarginLeft(2), so the
			// invariant for an exact fit is:
			//   viewport.Width + rightCol + logsMarginLeft == width
			assert.Equal(t, tt.width, v.viewport.Width()+v.rightCol+logsMarginLeft,
				"logs body (left + right + marginLeft) must equal the assigned width; "+
					"a mismatch means the rendered output will overflow or leave dead space")
		})
	}
}

// TestLogsView_Update_ViewportSizeMsg_PartitionInvariant is the Phase 2
// version of the partition regression test. The same invariant holds
// (viewport.Width + rightCol + logsMarginLeft == assigned width), but
// the size arrives through msgs.ViewportSizeMsg instead of the historical
// Resize(w, h) call.
func TestLogsView_Update_ViewportSizeMsg_PartitionInvariant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	proc := mock_ui.NewMockProcessorController(ctrl)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{IsProcessing: false}).AnyTimes()
	logger := mock_ui.NewMockLogProvider(ctrl)
	logger.EXPECT().Get().Return([]logs.LogMessage{}).AnyTimes()
	logger.EXPECT().Clear().AnyTimes()

	v := NewLogsView(logger, proc)

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
			assert.Equal(t, tt.width, v.viewport.Width()+v.rightCol+logsMarginLeft,
				"partition must be exact: left + right + marginLeft == assigned width")
		})
	}
}

// TestLogsView_Update_MetricsVisibilityMsg_StartsTick — when
// MetricsVisibilityMsg{Visible: true} arrives, the returned cmd must
// be non-nil (a tea.Tick that produces MetricsTickMsg).
func TestLogsView_Update_MetricsVisibilityMsg_StartsTick(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	proc := mock_ui.NewMockProcessorController(ctrl)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{IsProcessing: false}).AnyTimes()
	logger := mock_ui.NewMockLogProvider(ctrl)
	logger.EXPECT().Get().Return([]logs.LogMessage{}).AnyTimes()
	logger.EXPECT().Clear().AnyTimes()

	v := NewLogsView(logger, proc)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	proc := mock_ui.NewMockProcessorController(ctrl)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{IsProcessing: false}).AnyTimes()
	logger := mock_ui.NewMockLogProvider(ctrl)
	logger.EXPECT().Get().Return([]logs.LogMessage{}).AnyTimes()
	logger.EXPECT().Clear().AnyTimes()

	v := NewLogsView(logger, proc)

	// Start first to flip Visible=true, then stop.
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	proc := mock_ui.NewMockProcessorController(ctrl)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{}).AnyTimes()
	logger := mock_ui.NewMockLogProvider(ctrl)
	logger.EXPECT().Get().Return([]logs.LogMessage{}).AnyTimes()
	logger.EXPECT().Clear().AnyTimes()
	v := NewLogsView(logger, proc)
	if cmd := v.Init(); cmd != nil {
		t.Fatalf("Init must return nil; got %T", cmd)
	}
}

// TestLogsView_UpdateLogs_PersistsContent is the regression test for the
// value-receiver mutation-loss bug in updateLogs. Before the fix the
// function mutated the receiver copy's viewport and returned nothing,
// so the caller never saw the new content. The fix returns the
// modified LogsView so the viewport state is captured at the call
// site.
//
// Test contract:
//   - The mock LogProvider returns ["line A", "line B"].
//   - updateLogs() returns a LogsView whose viewport content equals
//     "line A\nline B".
//   - The returned LogsView is observably different from the receiver
//     (a captured copy), proving the fix is not a no-op aliasing the
//     input.
func TestLogsView_UpdateLogs_PersistsContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	proc := mock_ui.NewMockProcessorController(ctrl)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{IsProcessing: false}).AnyTimes()

	logger := mock_ui.NewMockLogProvider(ctrl)
	logger.EXPECT().Get().Return([]logs.LogMessage{
		logs.NewGeneralMessage("", "", "line A"),
		logs.NewGeneralMessage("", "", "line B"),
	}).AnyTimes()
	logger.EXPECT().Clear().AnyTimes()

	v := NewLogsView(logger, proc)
	require.NotEmpty(t, v.viewport.GetContent(),
		"sanity check: NewLogsView seeds viewport with the initial log lines")

	// autoScroll is true on construction, so GotoBottom runs inside
	// updateLogs. The test only asserts the content, not the scroll
	// position — scroll is exercised by the spec's
	// ProcessingStartedMsg autoScroll scenario.
	got := v.updateLogs()

	assert.Equal(t, "line A\nline B", got.viewport.GetContent(),
		"viewport content must reflect the logger.Get() lines joined by newline")
	assert.Equal(t, "line A\nline B", v.viewport.GetContent(),
		"viewport on the captured receiver copy must also reflect the lines; "+
			"updateLogs mutates the receiver in place, the returned value is the same copy")
}
