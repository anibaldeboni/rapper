package ui

// ChromeLayout is the single source of truth for chrome dimensions that
// frame the content area (global margins, header, status bar). It is a
// pure value type: zero value is invalid in production; instances are
// constructed exclusively via NewChromeLayout, which seeds the defaults
// that match the current TUI. Methods are O(1) integer arithmetic with
// no allocations — the AppModel's WindowSizeMsg handler runs only on
// tea.WindowSizeMsg.
//
// Future changes to the chrome (e.g. a taller status bar, an expanded
// header) only need to mutate this type: the AppModel re-reads the
// fields on the next resize and propagates the new dimensions to every
// view through the existing Resize chain.
type ChromeLayout struct {
	marginRows      int // top+bottom rows consumed by global Margin(1,2)
	marginCols      int // left+right columns consumed by global Margin(1,2)
	headerHeight    int // height of renderHeader() in rows
	statusBarHeight int // height of renderStatusBar() in rows
}

// NewChromeLayout returns a ChromeLayout populated with the production
// defaults: marginRows=2, marginCols=4, headerHeight=1, statusBarHeight=1.
// These values preserve the pre-change layout byte-for-byte (spec
// requirement 10) so swapping the magic -4 for a ChromeLayout call does
// not shift any view.
func NewChromeLayout() ChromeLayout {
	return ChromeLayout{
		marginRows:      2,
		marginCols:      4,
		headerHeight:    1,
		statusBarHeight: 1,
	}
}

// AvailableWidth returns the width available to views after chrome.
// No lower bound is applied here: the only consumer is the view Resize
// chain, which already handles non-positive widths gracefully.
func (c ChromeLayout) AvailableWidth(windowWidth int) int {
	return windowWidth - c.marginCols
}

// AvailableHeight returns the height available to views after chrome.
// Returns 0 or negative on degenerate input (window smaller than the
// chrome). Callers apply their own lower bound — the AppModel
// WindowSizeMsg handler, for example, uses max(..., 10) so views
// always have a usable area.
func (c ChromeLayout) AvailableHeight(windowHeight int) int {
	return windowHeight - c.marginRows - c.headerHeight - c.statusBarHeight
}
