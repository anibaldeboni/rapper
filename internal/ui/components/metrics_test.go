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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var _ tea.Msg = msgs.MetricsTickMsg(time.Now()) // keep import warm in case future tests need it

// newTestMetricsPanel builds a MetricsPanel with a gomock-backed
// ProcessorController. The default GetMetrics return is
// `ports.ProcessorMetrics{}` (zero value); tests that need richer metrics
// pass them via the variadic `metrics` argument. The expectation is
// registered as `.AnyTimes()` so tests can layer additional constraints
// by capturing the mock controller themselves.
func newTestMetricsPanel(t *testing.T, metrics ...ports.ProcessorMetrics) MetricsPanel {
	t.Helper()
	ctrl := gomock.NewController(t)
	proc := mock_ui.NewMockProcessorController(ctrl)
	m := ports.ProcessorMetrics{}
	if len(metrics) > 0 {
		m = metrics[0]
	}
	proc.EXPECT().GetMetrics().Return(m).AnyTimes()
	return NewMetricsPanel(proc)
}

func TestMetricsPanel_View_RendersAllMetricRows(t *testing.T) {
	p := newTestMetricsPanel(t, ports.ProcessorMetrics{
		TotalRequests:   10,
		SuccessRequests: 8,
		ErrorRequests:   2,
		LinesProcessed:  100,
		ActiveWorkers:   4,
		RequestsPerSec:  1.5,
		IsProcessing:    true,
	})

	out := p.View().Content

	// Each metric label must appear in the rendered output
	for _, label := range []string{"Status:", "Total Requests:", "✓ Success:", "✗ Errors:", "Lines Processed:", "Throughput:", "Active Workers:"} {
		assert.Contains(t, out, label, "metrics panel should show label %q", label)
	}
}

func TestMetricsPanel_Update_TickRefreshesMetrics(t *testing.T) {
	p := newTestMetricsPanel(t, ports.ProcessorMetrics{
		TotalRequests:  42,
		ActiveWorkers:  3,
		IsProcessing:   true,
		StartTime:      time.Now().Add(-2 * time.Second),
		RequestsPerSec: 21,
	})
	p = p.SetVisible(true)

	next, cmd := p.Update(msgs.MetricsTickMsg(time.Now()))
	p = next.(MetricsPanel)

	assert.NotNil(t, cmd, "tick should reschedule itself while visible")
	out := p.View().Content
	assert.Contains(t, out, "42", "view should show the total requests from the mock snapshot")
}

func TestMetricsPanel_SetVisible_StopsTickWhenHidden(t *testing.T) {
	p := newTestMetricsPanel(t)
	p = p.SetVisible(false)

	_, cmd := p.Update(msgs.MetricsTickMsg(time.Now()))

	assert.Nil(t, cmd, "tick should not reschedule when not visible")
}

// TestMetricsPanel_Update_MetricsVisibilityMsg_StartsTick — value-receiver
// tea.Model: Update(MetricsVisibilityMsg{Visible: true}) flips Visible
// and returns a tick cmd.
func TestMetricsPanel_Update_MetricsVisibilityMsg_StartsTick(t *testing.T) {
	p := newTestMetricsPanel(t)
	require.False(t, p.Visible, "panel starts hidden")

	next, cmd := p.Update(msgs.MetricsVisibilityMsg{Visible: true})
	p = next.(MetricsPanel)

	assert.True(t, p.Visible, "Visible must be true after Update(Visible: true)")
	assert.NotNil(t, cmd, "Update(Visible: true) must return a tick cmd")
}

// TestMetricsPanel_Update_MetricsVisibilityMsg_StopsTick — Update({Visible:
// false}) clears Visible and returns nil cmd.
func TestMetricsPanel_Update_MetricsVisibilityMsg_StopsTick(t *testing.T) {
	p := newTestMetricsPanel(t)
	p.Visible = true

	next, cmd := p.Update(msgs.MetricsVisibilityMsg{Visible: false})
	p = next.(MetricsPanel)

	assert.False(t, p.Visible, "Visible must be false after Update(Visible: false)")
	assert.Nil(t, cmd, "Update(Visible: false) must return nil cmd")
}

// TestMetricsPanel_Init_ReturnsNil — Init must return nil per R-6.
func TestMetricsPanel_Init_ReturnsNil(t *testing.T) {
	p := newTestMetricsPanel(t)
	if cmd := p.Init(); cmd != nil {
		t.Fatalf("Init must return nil; got %T", cmd)
	}
}

func TestMetricsPanel_View_ShowsIdleStatusWhenNotProcessing(t *testing.T) {
	p := newTestMetricsPanel(t, ports.ProcessorMetrics{IsProcessing: false})

	out := p.View().Content

	assert.Contains(t, out, "Idle", "idle status must be shown when not processing")
	assert.True(t, !strings.Contains(out, "🟢 Processing"), "processing indicator must not appear when idle")
}
