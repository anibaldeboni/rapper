package logo

import (
	"fmt"
	"strings"
	"time"

	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/ui/assets"
	"github.com/charmbracelet/lipgloss"
	"github.com/mazznoer/colorgrad"
)

type LogoOption func(*LogoConfiguration)

type LogoConfiguration struct {
	Style    *lipgloss.Style
	Image    string
	Gradient colorgrad.Gradient
}
type VerticalPosition int

const (
	Top VerticalPosition = iota
	Middle
	Bottom
)

// WithHorizontalPosition returns a LogoOption function that sets the position and width of the logo style.
// The position is set using the lipgloss.Position type, and the width is set using an integer value.
func WithHorizontalPosition(pos lipgloss.Position, width int) LogoOption {
	return func(config *LogoConfiguration) {
		*config.Style = config.Style.Width(width).Align(pos)
	}
}

func WithVerticalPosition(pos VerticalPosition, height int) LogoOption {
	return func(config *LogoConfiguration) {
		logoHeight := lipgloss.Height(config.Image)
		switch pos {
		case Top:
			*config.Style = config.Style.PaddingTop(0)
		case Middle:
			*config.Style = config.Style.PaddingTop((height - logoHeight) / 2)
		case Bottom:
			*config.Style = config.Style.PaddingBottom(height - logoHeight)
		}
	}
}

func WithCustomLogo(logo string) LogoOption {
	return func(config *LogoConfiguration) {
		config.Image = logo
	}
}

func WithMainLogo() LogoOption {
	return func(config *LogoConfiguration) {
		config.Image = assets.Logo()
	}
}

// WithRandomImage returns a LogoOption that sets the logo image to a random logo.
func WithRandomImage() LogoOption {
	return func(config *LogoConfiguration) {
		config.Image = assets.RandomLogo()
	}
}

// WithRandomGradient returns a LogoOption that sets the gradient of the logo configuration to a random color gradient.
func WithRandomGradient() LogoOption {
	return func(config *LogoConfiguration) {
		config.Gradient = randomColorGradient()
	}
}

func WithPinkToPurpleGradient() LogoOption {
	return func(config *LogoConfiguration) {
		config.Gradient = pinkToPurple
	}
}

// WithDefaultOptions returns a slice of LogoOption containing the default options for creating a logo.
// The default options include a random image and a random gradient.
func WithDefaultOptions() []LogoOption {
	return []LogoOption{
		WithRandomImage(),
		WithRandomGradient(),
	}
}

// Static returns the logo string with the specified configurations.
// The `configs` parameter is a variadic parameter that allows passing LogoOption values.
// These options can be used to customize the appearance of the logo.
// Returns the logo string.
func Static(configs ...LogoOption) string {
	return colorizeLogo(configs...)
}

// PrintAnimated prints the animated logo by splitting the colorized logo into lines
// and printing each line with a slight delay between them.
func PrintAnimated() {
	colorized := colorizeLogo(
		WithRandomImage(),
		WithRandomGradient(),
		WithHorizontalPosition(lipgloss.Center, styles.TerminalWidth()),
	)

	lines := strings.Split(colorized, "\n")

	fmt.Print("\n\n")

	for _, line := range lines {
		fmt.Println(line)
		time.Sleep(50 * time.Millisecond)
	}
}

func colorizeLogo(options ...LogoOption) string {
	config := LogoConfiguration{Style: &lipgloss.Style{}}
	for _, opt := range options {
		opt(&config)
	}

	splited := strings.Split(config.Image, "\n")
	lines := len(splited)
	grad := config.Gradient.Sharp(uint(lines), 0)
	steps := 1.0 / float64(lines)

	var colorized []string
	for i, line := range splited {
		color := grad.At(steps * float64(i))
		colorized = append(
			colorized,
			lipgloss.NewStyle().
				Foreground(lipgloss.Color(color.Hex())).
				Render(line),
		)
	}

	return config.Style.Render(strings.Join(colorized, "\n"))
}
