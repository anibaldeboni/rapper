package components

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	mock_ui "github.com/anibaldeboni/rapper/internal/ui/mock"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/anibaldeboni/rapper/internal/ui/ports"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMetricsPanel_View_RendersAllMetricRows(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	proc := mock_ui.NewMockProcessorController(ctrl)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{
		TotalRequests:   10,
		SuccessRequests: 8,
		ErrorRequests:   2,
		LinesProcessed:  100,
		ActiveWorkers:   4,
		RequestsPerSec:  1.5,
		IsProcessing:    true,
	}).AnyTimes()

	p := NewMetricsPanel(proc)

	out := p.View()

	// Each metric label must appear in the rendered output
	for _, label := range []string{"Status:", "Total Requests:", "✓ Success:", "✗ Errors:", "Lines Processed:", "Throughput:", "Active Workers:"} {
		assert.Contains(t, out, label, "metrics panel should show label %q", label)
	}
}

func TestMetricsPanel_Update_TickRefreshesMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	proc := mock_ui.NewMockProcessorController(ctrl)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{
		TotalRequests:  42,
		ActiveWorkers:  3,
		IsProcessing:   true,
		StartTime:      time.Now().Add(-2 * time.Second),
		RequestsPerSec: 21,
	}).AnyTimes()

	p := NewMetricsPanel(proc)
	p.SetVisible(true)

	var cmd tea.Cmd
	p, cmd = p.Update(msgs.MetricsTickMsg(time.Now()))

	assert.NotNil(t, cmd, "tick should reschedule itself while visible")
	out := p.View()
	assert.Contains(t, out, "42", "view should show the total requests from the mock snapshot")
}

func TestMetricsPanel_SetVisible_StopsTickWhenHidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	proc := mock_ui.NewMockProcessorController(ctrl)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{}).AnyTimes()

	p := NewMetricsPanel(proc)
	p.SetVisible(false)

	_, cmd := p.Update(msgs.MetricsTickMsg(time.Now()))

	assert.Nil(t, cmd, "tick should not reschedule when not visible")
}

func TestMetricsPanel_View_ShowsIdleStatusWhenNotProcessing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	proc := mock_ui.NewMockProcessorController(ctrl)
	proc.EXPECT().GetMetrics().Return(ports.ProcessorMetrics{IsProcessing: false}).AnyTimes()

	p := NewMetricsPanel(proc)

	out := p.View()

	assert.Contains(t, out, "Idle", "idle status must be shown when not processing")
	assert.True(t, !strings.Contains(out, "🟢 Processing"), "processing indicator must not appear when idle")
}
