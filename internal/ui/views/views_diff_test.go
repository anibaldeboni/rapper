package views

import (
	"os"
	"regexp"
	"sort"
	"testing"
)

type viewKeyHandlingGolden struct {
	count int
	keys  []string
}

func TestViews_UpdateKeyHandlingArms_NoChange(t *testing.T) {
	golden := map[string]viewKeyHandlingGolden{
		"files.go": {count: 1, keys: []string{"Select"}},
		"logs.go": {
			count: 7,
			keys:  []string{"Down", "GotoBottom", "GotoTop", "PageDown", "PageUp", "Select", "Up"},
		},
		"settings.go": {count: 4, keys: []string{"Edit", "ExitEdit", "Save", "Select"}},
	}

	matcher := regexp.MustCompile(`key\.Matches\(msg,\s*kbind\.(\w+)`)
	for file, want := range golden {
		t.Run(file, func(t *testing.T) {
			source, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("read %s: %v", file, err)
			}
			matches := matcher.FindAllStringSubmatch(string(source), -1)
			gotKeys := make([]string, len(matches))
			for i, match := range matches {
				gotKeys[i] = match[1]
			}
			sort.Strings(gotKeys)

			if len(matches) != want.count {
				t.Fatalf("%s key handling arm count = %d; want %d", file, len(matches), want.count)
			}
			sort.Strings(want.keys)
			if got := gotKeys; !equalStrings(got, want.keys) {
				t.Fatalf("%s key handling arms = %v; want %v", file, got, want.keys)
			}
		})
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
