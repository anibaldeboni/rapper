package views

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func bindingKeys(bindings []key.Binding) [][]string {
	keys := make([][]string, len(bindings))
	for i, binding := range bindings {
		keys[i] = binding.Keys()
	}
	return keys
}

func bindingGroupKeys(groups [][]key.Binding) [][][]string {
	keys := make([][][]string, len(groups))
	for i, group := range groups {
		keys[i] = bindingKeys(group)
	}
	return keys
}

func TestFilesView_KeyMap_ShortHelp_MirrorsPriorCentralized(t *testing.T) {
	got := bindingKeys(NewFilesView(nil, ',').KeyMap().ShortHelp())
	want := [][]string{{"up"}, {"down"}, {"enter"}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FilesView.KeyMap().ShortHelp() keys = %v; want %v", got, want)
	}

	gotFull := bindingGroupKeys(NewFilesView(nil, ',').KeyMap().FullHelp())
	wantFull := [][][]string{{{"up"}, {"down"}, {"enter"}}}
	if !reflect.DeepEqual(gotFull, wantFull) {
		t.Fatalf("FilesView.KeyMap().FullHelp() keys = %v; want %v", gotFull, wantFull)
	}
}

func TestFilesView_Update_ViewActivatedMsg_EmitsStatusKeymap(t *testing.T) {
	v := NewFilesView(nil, ',')
	next, cmd := v.Update(msgs.ViewActivatedMsg{})
	if _, ok := next.(FilesView); !ok {
		t.Fatalf("Update(ViewActivatedMsg) returned %T; want FilesView", next)
	}
	if cmd == nil {
		t.Fatal("Update(ViewActivatedMsg) must return a status-keymap command")
	}
	status, ok := cmd().(msgs.StatusKeymapMsg)
	if !ok {
		t.Fatalf("activation command returned %T; want msgs.StatusKeymapMsg", cmd())
	}
	if got, want := bindingKeys(status.Current.ShortHelp()), bindingKeys(v.KeyMap().ShortHelp()); !reflect.DeepEqual(got, want) {
		t.Fatalf("activation keymap keys = %v; want %v", got, want)
	}
}

// TestFilesView_Init_ReturnsNil — Init must return nil per R-6.
func TestFilesView_Init_ReturnsNil(t *testing.T) {
	v := NewFilesView(nil, ',')
	cmd := v.Init()
	if cmd != nil {
		t.Fatalf("Init must return nil; got %T", cmd)
	}
}

// TestFilesView_Update_SelectEmitsItemSelectedMsg — pressing Select on a
// focused file must emit ItemSelectedMsg{FilePath} (R-8).
func TestFilesView_Update_SelectEmitsItemSelectedMsg(t *testing.T) {
	items := []list.Item{
		Option[string]{Value: "a.csv", Title: "a.csv"},
		Option[string]{Value: "b.csv", Title: "b.csv"},
	}
	v := NewFilesView(items, ',')
	v.list.Select(0) // focus first item

	// Synthesize the key message the bubbles list delivers on Select.
	keyMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	next, cmd := v.Update(keyMsg)
	_ = next

	if cmd == nil {
		t.Fatal("Update(Select) must return a non-nil cmd that yields ItemSelectedMsg")
	}
	out := cmd()
	selected, ok := out.(msgs.ItemSelectedMsg)
	if !ok {
		t.Fatalf("cmd must yield msgs.ItemSelectedMsg; got %T", out)
	}
	if selected.FilePath != "a.csv" {
		t.Fatalf("ItemSelectedMsg.FilePath = %q; want %q", selected.FilePath, "a.csv")
	}
}

// TestFilesView_Update_ViewportSizeMsg_ResizesList — Update(ViewportSizeMsg)
// must partition the available width between the list and preview table.
func TestFilesView_Update_ViewportSizeMsg_ResizesList(t *testing.T) {
	v := NewFilesView(nil, ',')

	cases := []struct {
		name                         string
		w, h                         int
		wantListW, wantTableW, wantH int
	}{
		{"zero 0×20 → list 0, table 0", 0, 20, 0, 0, 20},
		{"very narrow 1×20 → list 0, table 0", 1, 20, 0, 0, 20},
		{"narrow 4×20 → list 1, table 2", 4, 20, 1, 2, 20},
		{"common 80×20 → list 20, table 59", 80, 20, 20, 59, 20},
		{"wide 120×30 → list 30, table 89", 120, 30, 30, 89, 30},
		{"approved 160×30 → list 40, table 119", 160, 30, 40, 119, 30},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			next, _ := v.Update(msgs.ViewportSizeMsg{Width: tc.w, Height: tc.h})
			nextFiles, ok := next.(FilesView)
			if !ok {
				t.Fatalf("Update must return FilesView value; got %T", next)
			}
			if got := nextFiles.list.Width(); got != tc.wantListW {
				t.Errorf("list.Width() = %d; want %d", got, tc.wantListW)
			}
			if got := nextFiles.table.Width(); got != tc.wantTableW {
				t.Errorf("table.Width() = %d; want %d", got, tc.wantTableW)
			}
			if tc.w > 0 {
				assert.LessOrEqual(t, nextFiles.list.Width(), tc.w/4)
				assert.Equal(t, tc.w, nextFiles.list.Width()+nextFiles.table.Width()+previewPaneMarginLeft)
			}
			if got := nextFiles.list.Height(); got != tc.wantH {
				t.Errorf("list.Height() = %d; want %d", got, tc.wantH)
			}
		})
	}
}

func TestFilesView_PreviewUsesSharedTitleStyle(t *testing.T) {
	path := writePreviewCSV(t, "id,name\n1,Ada\n")
	v := NewFilesView([]list.Item{Option[string]{Value: path, Title: "A"}}, ',')
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 40, Height: 10})

	content := next.(FilesView).View().Content
	require.Contains(t, content, styles.TitleStyle.Render("📄 Preview"))
}

func TestPreviewTableStyles_LeftAlignAndMatchPadding(t *testing.T) {
	got := previewTableStyles()

	require.Equal(t, lipgloss.Left, got.Header.GetAlignHorizontal())
	require.Equal(t, lipgloss.Left, got.Cell.GetAlignHorizontal())
	require.Equal(t, got.Header.GetPaddingLeft(), got.Cell.GetPaddingLeft())
	require.Equal(t, got.Header.GetPaddingRight(), got.Cell.GetPaddingRight())
}

func TestFilesView_PaginationDots_AreByteIdentical(t *testing.T) {
	v := NewFilesView(nil, ',')

	tests := []struct {
		name  string
		style interface{ Render(...string) string }
		want  string
	}{
		{
			name:  "active",
			style: v.list.Styles.ActivePaginationDot,
			want:  "\x1b[1;38;2;211;211;211m⦿ \x1b[m",
		},
		{
			name:  "inactive",
			style: v.list.Styles.InactivePaginationDot,
			want:  "\x1b[1;38;2;141;141;141m⦁ \x1b[m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.style.Render(""); got != tt.want {
				t.Fatalf("pagination dot = %q; want %q", got, tt.want)
			}
		})
	}
}

// TestFilesView_View_ReturnsTeaView — View() must return tea.View
// (R-28 / R-16) whose Content matches the underlying list.
func TestFilesView_View_ReturnsTeaView(t *testing.T) {
	v := NewFilesView(nil, ',')
	out := v.View()
	if out.Content == "" {
		t.Fatal("View().Content must not be empty (list always renders something)")
	}
}

// TestFilesView_ValueReceiver_PreservesState — calling Update must
// mutate the value receiver's copy AND the caller must capture the
// returned next. Without capture, state is silently lost. This test
// pins the contract by proving that the next-value has the new
// dimensions and the original value does not (since it was never
// mutated).
//
// We resize to 80x20 and assert that:
//   - next (the returned value) has the new dimensions
//   - the captured original v has the constructor defaults (60x5)
//
// The contrast proves the value receiver is in effect: the
// constructor-time dimensions survive on the original because the
// value was copied, not mutated.
func TestFilesView_ValueReceiver_PreservesState(t *testing.T) {
	v := NewFilesView(nil, ',')
	original := v

	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})

	// Original v kept its constructor dimensions (60x5), proving
	// the value receiver did not mutate the caller's copy.
	if v.list.Width() != 60 || v.list.Height() != 5 {
		t.Errorf("original FilesView must keep constructor dims (60x5); got %dx%d",
			v.list.Width(), v.list.Height())
	}

	// next reflects the resize (80 → list 20, table 59, 20).
	nextFiles := next.(FilesView)
	if nextFiles.list.Width() != 20 || nextFiles.list.Height() != 20 || nextFiles.table.Width() != 59 {
		t.Errorf("next FilesView must reflect Update; got list=%dx%d table=%d; want list=20x20 table=59",
			nextFiles.list.Width(), nextFiles.list.Height(), nextFiles.table.Width())
	}

	_ = original
}

func TestFilesView_SelectionChangeRefreshesPreview(t *testing.T) {
	aPath := writePreviewCSV(t, "id,name\n1,Ada\n")
	bPath := writePreviewCSV(t, "id,name\n2,Bea\n")
	v := NewFilesView([]list.Item{
		Option[string]{Value: aPath, Title: "A"},
		Option[string]{Value: bPath, Title: "B"},
	}, ',')

	requirePreviewRows(t, v, [][]string{{"1", "Ada"}})

	next, _ := v.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	nextFiles := next.(FilesView)

	requirePreviewRows(t, nextFiles, [][]string{{"2", "Bea"}})
}

func TestFilesView_UnchangedSelectionDoesNotReloadPreview(t *testing.T) {
	path := writePreviewCSV(t, "id\noriginal\n")
	v := NewFilesView([]list.Item{Option[string]{Value: path, Title: "A"}}, ',')
	requirePreviewRows(t, v, [][]string{{"original"}})

	require.NoError(t, os.WriteFile(path, []byte("id\nupdated\n"), 0o600))
	next, _ := v.Update(tea.KeyPressMsg{Code: tea.KeyUp})

	requirePreviewRows(t, next.(FilesView), [][]string{{"original"}})
}

func TestFilesView_PreviewIsUnfocusedAndListOwnsKeys(t *testing.T) {
	path := writePreviewCSV(t, "id\n1\n")
	v := NewFilesView([]list.Item{Option[string]{Value: path, Title: "A"}}, ',')

	if v.table.Focused() {
		t.Fatal("preview table must remain unfocused")
	}

	next, cmd := v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)
	assert.Equal(t, path, cmd().(msgs.ItemSelectedMsg).FilePath)
	assert.False(t, next.(FilesView).table.Focused())
}

func TestFilesView_PreviewErrorDoesNotDisableSelection(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.csv")
	validPath := writePreviewCSV(t, "id\n1\n")
	v := NewFilesView([]list.Item{
		Option[string]{Value: missingPath, Title: "Missing"},
		Option[string]{Value: validPath, Title: "Valid"},
	}, ',')

	if v.previewErr == "" {
		t.Fatal("missing preview file must produce a local preview error")
	}

	next, _ := v.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	nextFiles := next.(FilesView)
	assert.Empty(t, nextFiles.previewErr)
	requirePreviewRows(t, nextFiles, [][]string{{"1"}})
}

func TestFilesView_WithSeparatorRefreshesSelectedPreview(t *testing.T) {
	path := writePreviewCSV(t, "id;name\n1;Ada\n")
	v := NewFilesView([]list.Item{Option[string]{Value: path, Title: "A"}}, ',')
	assert.Len(t, v.table.Columns(), 1)

	next := v.WithSeparator(';')
	assert.False(t, next.table.Focused())
	assert.Equal(t, []string{"id", "name"}, tableColumnTitles(next))
	requirePreviewRows(t, next, [][]string{{"1", "Ada"}})
}

func TestFilesView_NarrowLayoutAccountsForPreviewMargin(t *testing.T) {
	path := writePreviewCSV(t, "id,name\n1,Ada\n")
	v := NewFilesView([]list.Item{Option[string]{Value: path, Title: "A"}}, ',')
	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 18, Height: 10})
	nextFiles := next.(FilesView)

	require.Equal(t, 4, nextFiles.list.Width())
	require.Equal(t, 13, nextFiles.table.Width())
	require.Equal(t, 18, nextFiles.list.Width()+nextFiles.table.Width()+previewPaneMarginLeft)

	for _, line := range strings.Split(nextFiles.View().Content, "\n") {
		if width := lipgloss.Width(line); width > 18 {
			t.Fatalf("line width = %d; line = %q", width, line)
		}
	}
}

func writePreviewCSV(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "preview.csv")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func requirePreviewRows(t *testing.T, v FilesView, want [][]string) {
	t.Helper()
	rows := make([][]string, len(v.table.Rows()))
	for i, row := range v.table.Rows() {
		rows[i] = []string(row)
	}
	require.Equal(t, want, rows)
}

func tableColumnTitles(v FilesView) []string {
	titles := make([]string, len(v.table.Columns()))
	for i, column := range v.table.Columns() {
		titles[i] = column.Title
	}
	return titles
}
