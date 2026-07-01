package views

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/ui/components"
	mock_ui "github.com/anibaldeboni/rapper/internal/ui/mock"
	"github.com/anibaldeboni/rapper/internal/ui/ports"
	"github.com/stretchr/testify/assert"
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
	logger.EXPECT().Get().Return([]string{}).AnyTimes()

	v := NewLogsView(logger, proc)
	v.Resize(120, 40)

	// The metrics panel must be at least as wide as its longest rendered
	// row, otherwise lipgloss clips the right edge of every row that
	// exceeds the column width (the user reported "está cortando os
	// textos" — the text is being cut).
	requiredWidth := longestVisibleLine(v.metrics.View())
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
	logger.EXPECT().Get().Return([]string{}).AnyTimes()

	v := NewLogsView(logger, proc)
	v.Resize(120, 40)

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
	logger.EXPECT().Get().Return([]string{}).AnyTimes()

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
			v.Resize(tt.width, tt.height)

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
