package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// chromeDefaults returns a ChromeLayout populated with the production
// defaults: marginRows=2, marginCols=4, headerHeight=1, statusBarHeight=1.
// Centralised so every test case that wants a "fresh" chrome reuses the
// same baseline; only the field under test is overridden.
func chromeDefaults() ChromeLayout {
	return ChromeLayout{
		marginRows:      2,
		marginCols:      4,
		headerHeight:    1,
		statusBarHeight: 1,
	}
}

func TestChromeLayout_AvailableWidth_Defaults(t *testing.T) {
	// Spec S1: defaults minus marginCols (4) yields 116 for a 120-wide
	// terminal. The chrome-deduction constant must NOT appear in the
	// production code — these tests are the contract that the value
	// remains correct without magic numbers.
	tests := []struct {
		name        string
		windowWidth int
		want        int
	}{
		{name: "spec S1 — 120 wide yields 116", windowWidth: 120, want: 116},
		{name: "common 80-col terminal", windowWidth: 80, want: 76},
		{name: "narrow 40-col terminal", windowWidth: 40, want: 36},
		{name: "wide 200-col terminal", windowWidth: 200, want: 196},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := chromeDefaults()
			assert.Equal(t, tt.want, c.AvailableWidth(tt.windowWidth))
		})
	}
}

func TestChromeLayout_AvailableHeight_Defaults(t *testing.T) {
	// Spec S1: defaults deduct marginRows (2) + headerHeight (1) +
	// statusBarHeight (1) = 4 rows total, so a 40-row terminal gives 36.
	tests := []struct {
		name         string
		windowHeight int
		want         int
	}{
		{name: "spec S1 — 40 tall yields 36", windowHeight: 40, want: 36},
		{name: "common 24-row terminal", windowHeight: 24, want: 20},
		{name: "tall 80-row terminal", windowHeight: 80, want: 76},
		{name: "minimum 10-row terminal", windowHeight: 10, want: 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := chromeDefaults()
			assert.Equal(t, tt.want, c.AvailableHeight(tt.windowHeight))
		})
	}
}

func TestChromeLayout_AvailableHeight_HeaderChange(t *testing.T) {
	// Spec S6: raising headerHeight by 1 row (1 -> 2) subtracts one more
	// row from the available height. The test exercises several
	// terminal sizes to make sure the deduction scales linearly.
	tests := []struct {
		name         string
		windowHeight int
		headerHeight int
		want         int
	}{
		{name: "spec S6 — 40 tall with header=2 yields 35", windowHeight: 40, headerHeight: 2, want: 35},
		{name: "24 tall with header=2 yields 19", windowHeight: 24, headerHeight: 2, want: 19},
		{name: "80 tall with header=3 yields 74", windowHeight: 80, headerHeight: 3, want: 74},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := chromeDefaults()
			c.headerHeight = tt.headerHeight
			assert.Equal(t, tt.want, c.AvailableHeight(tt.windowHeight))
		})
	}
}

func TestChromeLayout_AvailableHeight_StatusBarChange(t *testing.T) {
	// Spec S7: raising statusBarHeight by 2 rows (1 -> 3) subtracts
	// two more rows from the available height. Tests cover several
	// sizes to make the deduction explicit.
	tests := []struct {
		name            string
		windowHeight    int
		statusBarHeight int
		want            int
	}{
		{name: "spec S7 — 40 tall with status=3 yields 34", windowHeight: 40, statusBarHeight: 3, want: 34},
		{name: "24 tall with status=3 yields 18", windowHeight: 24, statusBarHeight: 3, want: 18},
		{name: "80 tall with status=5 yields 72", windowHeight: 80, statusBarHeight: 5, want: 72},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := chromeDefaults()
			c.statusBarHeight = tt.statusBarHeight
			assert.Equal(t, tt.want, c.AvailableHeight(tt.windowHeight))
		})
	}
}

func TestChromeLayout_AvailableHeight_DegenerateInput(t *testing.T) {
	// Spec requirement 4: callers apply their own lower bound. The
	// layout type itself is allowed to return non-positive values on
	// pathological input — the contract is "max(windowHeight - chrome, 0)"
	// semantics, which means 0 or negative on a window smaller than the
	// chrome (a test-only scenario; production always passes a real
	// tea.WindowSizeMsg with positive dimensions).
	tests := []struct {
		name         string
		windowHeight int
		want         int
	}{
		{name: "zero-height window yields -4", windowHeight: 0, want: -4},
		{name: "height below chrome (3) yields -1", windowHeight: 3, want: -1},
		{name: "height equal to chrome (4) yields 0", windowHeight: 4, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := chromeDefaults()
			assert.Equal(t, tt.want, c.AvailableHeight(tt.windowHeight))
		})
	}
}

func TestChromeLayout_AvailableWidth_DegenerateInput(t *testing.T) {
	// Spec requirement 5: AvailableWidth returns windowWidth - marginCols.
	// No lower bound is applied; the caller (broadcastResize) does not
	// need one because the only consumer is the view Resize chain which
	// handles non-positive widths by leaving the view at its previous
	// state. This test pins the contract for the zero / negative case so
	// a future "defensive clamp" cannot silently change the result.
	tests := []struct {
		name        string
		windowWidth int
		want        int
	}{
		{name: "zero-width window yields -4", windowWidth: 0, want: -4},
		{name: "width below marginCols yields negative", windowWidth: 3, want: -1},
		{name: "width equal to marginCols yields 0", windowWidth: 4, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := chromeDefaults()
			assert.Equal(t, tt.want, c.AvailableWidth(tt.windowWidth))
		})
	}
}

func TestChromeLayout_ValueReceiver_DoesNotMutateOriginal(t *testing.T) {
	// Spec requirement 1: ChromeLayout is a value type, not a pointer.
	// Method calls on a value receiver must NOT mutate the original.
	// This guards against a future refactor that promotes the methods
	// to pointer receivers and silently introduces mutation.
	c := chromeDefaults()
	orig := c

	_ = c.AvailableWidth(120)
	_ = c.AvailableHeight(40)

	assert.Equal(t, orig, c, "ChromeLayout value receiver must not mutate the original")
}
