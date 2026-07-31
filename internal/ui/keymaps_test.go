package ui

import (
	"os"
	"strings"
	"testing"
)

func readKeymapsSource(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile("keymaps.go")
	if err != nil {
		t.Fatalf("read keymaps.go: %v", err)
	}
	return string(source)
}

func TestKeymaps_File_HasNoPerViewTypes(t *testing.T) {
	source := readKeymapsSource(t)
	for _, symbol := range []string{"type filesViewKeyMap", "type logsViewKeyMap"} {
		if strings.Contains(source, symbol) {
			t.Errorf("keymaps.go must not define %q", symbol)
		}
	}
}

func TestKeymaps_File_NoGetViewSpecificKeyMap(t *testing.T) {
	if strings.Contains(readKeymapsSource(t), "getViewSpecificKeyMap") {
		t.Fatal("keymaps.go must not define getViewSpecificKeyMap")
	}
}

func TestKeymaps_File_RetainsGlobalKeyMap(t *testing.T) {
	source := readKeymapsSource(t)
	for _, symbol := range []string{"type globalKeyMap", "func getGlobalKeyMap"} {
		if !strings.Contains(source, symbol) {
			t.Errorf("keymaps.go must retain %q", symbol)
		}
	}
}
