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
