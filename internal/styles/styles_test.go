package styles

import "testing"

func TestStyles_DarkRendering_IsByteIdentical(t *testing.T) {
	tests := []struct {
		name  string
		style interface{ Render(...string) string }
		input string
		want  string
	}{
		{name: "help key", style: HelpKeyStyle, input: "KEY", want: "\x1b[1;38;2;211;211;211mKEY\x1b[m"},
		{name: "help description", style: HelpDescStyle, input: "description", want: "\x1b[38;2;141;141;141mdescription\x1b[m"},
		{name: "help separator", style: HelpSepStyle, input: " • ", want: "\x1b[38;2;83;83;83m • \x1b[m"},
		{name: "active pagination dot", style: ActivePaginationDot, input: "", want: "\x1b[1;38;2;211;211;211m⦿ \x1b[m"},
		{name: "inactive pagination dot", style: InactivePaginationDot, input: "", want: "\x1b[1;38;2;141;141;141m⦁ \x1b[m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.style.Render(tt.input)
			t.Logf("rendered: %q", got)
			if got != tt.want {
				t.Fatalf("rendered output = %q; want %q", got, tt.want)
			}
		})
	}
}
