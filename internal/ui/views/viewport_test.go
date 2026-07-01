package views

import (
	"fmt"
	"testing"

	"charm.land/bubbles/v2/viewport"
)

// testView is a minimal repro of the value-receiver mutation-loss
// pattern: the viewport field is a value (not a pointer) inside a
// value-receiver struct, so an `updateLogs` helper that mutates v.vp
// loses the change unless the helper returns the modified copy and
// the caller captures it. The elm-messaging-logs-fix slice applies
// the same pattern to LogsView.updateLogs; this test verifies the
// pattern works in isolation.
type testView struct {
	vp viewport.Model
}

// updateLogs returns the modified testView so the caller can capture
// the new viewport state.
func (v testView) updateLogs() testView {
	content := "hello world"
	v.vp.SetContent(content)
	fmt.Println("After SetContent in updateLogs:", v.vp.GetContent())
	return v
}

func (v testView) Update() testView {
	v = v.updateLogs()
	fmt.Println("After updateLogs in Update:", v.vp.GetContent())
	vp2, _ := v.vp.Update(nil)
	fmt.Println("After vp.Update(nil):", vp2.GetContent())
	v.vp = vp2
	fmt.Println("After assignment:", v.vp.GetContent())
	return v
}

func TestViewportSetContentViaStruct(t *testing.T) {
	v := testView{vp: viewport.New(viewport.WithWidth(10), viewport.WithHeight(5))}
	v = v.Update()
	fmt.Println("After Update:", v.vp.GetContent())
	if v.vp.GetContent() == "" {
		t.Error("GetContent should not be empty")
	}
}
