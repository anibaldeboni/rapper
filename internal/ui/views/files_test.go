package views

import (
	"reflect"
	"testing"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/anibaldeboni/rapper/internal/ui/msgs"
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
	got := bindingKeys(NewFilesView(nil).KeyMap().ShortHelp())
	want := [][]string{{"up"}, {"down"}, {"enter"}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FilesView.KeyMap().ShortHelp() keys = %v; want %v", got, want)
	}

	gotFull := bindingGroupKeys(NewFilesView(nil).KeyMap().FullHelp())
	wantFull := [][][]string{{{"up"}, {"down"}, {"enter"}}}
	if !reflect.DeepEqual(gotFull, wantFull) {
		t.Fatalf("FilesView.KeyMap().FullHelp() keys = %v; want %v", gotFull, wantFull)
	}
}

func TestFilesView_Update_ViewActivatedMsg_EmitsStatusKeymap(t *testing.T) {
	v := NewFilesView(nil)
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
	v := NewFilesView(nil)
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
	v := NewFilesView(items)
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
// must call list.SetSize((w/4)*3, h) per R-9 / I-4.
func TestFilesView_Update_ViewportSizeMsg_ResizesList(t *testing.T) {
	v := NewFilesView(nil)

	cases := []struct {
		name         string
		w, h         int
		wantW, wantH int
	}{
		{"common 80×20 → list 60×20", 80, 20, 60, 20},
		{"narrow 40×20 → list 30×20", 40, 20, 30, 20},
		{"wide 120×30 → list 90×30", 120, 30, 90, 30},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			next, _ := v.Update(msgs.ViewportSizeMsg{Width: tc.w, Height: tc.h})
			nextFiles, ok := next.(FilesView)
			if !ok {
				t.Fatalf("Update must return FilesView value; got %T", next)
			}
			if got := nextFiles.list.Width(); got != tc.wantW {
				t.Errorf("list.Width() = %d; want %d", got, tc.wantW)
			}
			if got := nextFiles.list.Height(); got != tc.wantH {
				t.Errorf("list.Height() = %d; want %d", got, tc.wantH)
			}
		})
	}
}

func TestFilesView_PaginationDots_AreByteIdentical(t *testing.T) {
	v := NewFilesView(nil)

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
	v := NewFilesView(nil)
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
	v := NewFilesView(nil)
	original := v

	next, _ := v.Update(msgs.ViewportSizeMsg{Width: 80, Height: 20})

	// Original v kept its constructor dimensions (60x5), proving
	// the value receiver did not mutate the caller's copy.
	if v.list.Width() != 60 || v.list.Height() != 5 {
		t.Errorf("original FilesView must keep constructor dims (60x5); got %dx%d",
			v.list.Width(), v.list.Height())
	}

	// next reflects the resize (80 → 60, 20).
	nextFiles := next.(FilesView)
	if nextFiles.list.Width() != 60 || nextFiles.list.Height() != 20 {
		t.Errorf("next FilesView must reflect Update; got %dx%d; want 60x20",
			nextFiles.list.Width(), nextFiles.list.Height())
	}

	_ = original
}
