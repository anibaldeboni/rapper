package logo

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/mazznoer/colorgrad"
	"github.com/michaelquigley/figlet/figletlib"
)

// WithRandomFont returns a LogoOption that sets a random font for the logo.
func WithRandomFont() LogoOption {
	return func(config *LogoConfiguration) {
		config.Font = randomFiglet()
	}
}

// WithFont is a function that returns a LogoOption function which sets the font for the logo.
// It takes a font string as input and returns a LogoOption function that sets the font in the LogoConfiguration.
// If the font is not found, it panics with an error message.
// If there is an error reading the font, it panics with the error.
func WithFont(font string) LogoOption {
	figlet := figlets[font]

	if figlet == nil {
		panic(fmt.Sprintf("font %s not found", font))
	}

	figletFont, err := figletlib.ReadFontFromBytes(figlet)
	if err != nil {
		panic(err)
	}
	return func(config *LogoConfiguration) {
		config.Font = figletFont
	}
}

// WithRandomGradient returns a LogoOption that sets the gradient of the logo configuration to a random color gradient.
func WithRandomGradient() LogoOption {
	return func(config *LogoConfiguration) {
		config.Gradient = randomColorGradient()
	}
}

// WithGradient sets the gradient for the logo.
func WithGradient(grad colorgrad.Gradient) LogoOption {
	return func(config *LogoConfiguration) {
		config.Gradient = grad
	}
}

// WithHorizontalPosition returns a LogoOption function that sets the position and width of the logo style.
// The position is set using the lipgloss.Position type, and the width is set using an integer value.
func WithHorizontalPosition(pos lipgloss.Position, width int) LogoOption {
	return func(config *LogoConfiguration) {
		*config.Style = config.Style.Width(width).Align(pos)
	}
}

// WithCustomText is a LogoOption function that sets the custom text for the logo.
func WithCustomText(text string) LogoOption {
	return func(config *LogoConfiguration) {
		config.Text = text
	}
}
