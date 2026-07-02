package ui

import (
	"testing"

	mock_ui "github.com/anibaldeboni/rapper/internal/ui/mock"
)

// NewTestApp is the black-box re-export of newTestApp for tests living in
// package ui_test (or any other _test package). The file's `_test.go`
// suffix means it is compiled only by `go test` and never ships in the
// production binary.
//
// csvPaths is variadic: a zero-arg call preserves the legacy behavior
// (NewApp([]string{}, ...)) and a non-empty call passes the supplied
// paths verbatim to NewApp. All 7 default EXPECT().AnyTimes() calls are
// registered on the returned mocks.
func NewTestApp(t *testing.T, csvPaths ...string) (
	*AppModel,
	*mock_ui.MockLogService,
	*mock_ui.MockConfigManager,
	*mock_ui.MockProcessorController,
) {
	return newTestApp(t, csvPaths...)
}
