package logs

import (
	"fmt"

	"github.com/anibaldeboni/rapper/internal/styles"
)

var _ Message = (*messageImpl)(nil)

type Message interface {
	String() string
	WithIcon(string) Message
	WithKind(string) Message
	WithMessage(string) Message
}

type messageImpl struct {
	message string
	kind    string
	icon    string
}

func (m *messageImpl) String() string {
	var icon, kind string
	if m.kind != "" {
		kind = fmt.Sprintf("[%s] ", styles.Bold(m.kind))
	}
	if m.icon != "" {
		icon = m.icon + " "
	}
	return icon + kind + m.message
}

func (m *messageImpl) WithIcon(icon string) Message {
	m.icon = icon
	return m
}

func (m *messageImpl) WithKind(kind string) Message {
	m.kind = kind
	return m
}

func (m *messageImpl) WithMessage(message string) Message {
	m.message = message
	return m
}

func NewMessage() Message {
	return &messageImpl{}
}
