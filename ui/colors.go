package ui

import (
	"fmt"

	"github.com/fatih/color"
)

func toString(s any) string {
	var str string
	switch s.(type) {
	case int:
		str = fmt.Sprintf("%d", s)
		break
	default:
		return s.(string)
	}
	return str
}

func Bold(s any) string {
	b := color.New(color.FgWhite, color.Bold).SprintFunc()
	return b(toString(s))
}
func Italic(s any) string {
	i := color.New(color.FgWhite, color.Italic).SprintFunc()
	return i(toString(s))
}
func Red(s any) string {
	r := color.New(color.FgRed).SprintFunc()
	return r(toString(s))
}
