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

func (this *messageImpl) String() string {
	var icon, kind string
	if this.kind != "" {
		kind = fmt.Sprintf("[%s] ", styles.Bold(this.kind))
	}
	if this.icon != "" {
		icon = this.icon + " "
	}
	return icon + kind + this.message
}

func (this *messageImpl) WithIcon(icon string) Message {
	this.icon = icon
	return this
}

func (this *messageImpl) WithKind(kind string) Message {
	this.kind = kind
	return this
}

func (this *messageImpl) WithMessage(message string) Message {
	this.message = message
	return this
}

func NewMessage() Message {
	return &messageImpl{}
}
