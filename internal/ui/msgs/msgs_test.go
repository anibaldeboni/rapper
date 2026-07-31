package msgs

import (
	"reflect"
	"testing"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"github.com/stretchr/testify/assert"
)

type testKeyMap struct {
	name string
}

func (testKeyMap) ShortHelp() []key.Binding  { return nil }
func (testKeyMap) FullHelp() [][]key.Binding { return nil }

func TestMsgs_StatusKeymapMsg_CarriesCurrent(t *testing.T) {
	tests := []struct {
		name   string
		keyMap help.KeyMap
	}{
		{name: "empty map", keyMap: testKeyMap{}},
		{name: "named map", keyMap: testKeyMap{name: "logs"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := StatusKeymapMsg{Current: tt.keyMap}

			assert.Equal(t, tt.keyMap, msg.Current)
		})
	}
}

func TestMsgs_ViewActivatedMsg_IsEmpty(t *testing.T) {
	msg := ViewActivatedMsg{}

	assert.Equal(t, 0, reflect.TypeOf(msg).NumField())
}
