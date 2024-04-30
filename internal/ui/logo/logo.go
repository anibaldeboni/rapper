package logo

import (
	"fmt"
	"strings"
	"time"

	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/ui"
	"github.com/anibaldeboni/rapper/internal/ui/assets"
	"github.com/anibaldeboni/rapper/internal/utils"
	"github.com/charmbracelet/lipgloss"
	"github.com/mazznoer/colorgrad"
	"github.com/michaelquigley/figlet/figletlib"
	"golang.org/x/exp/maps"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type LogoOption func(*LogoConfiguration)

type LogoConfiguration struct {
	Style    *lipgloss.Style
	Gradient colorgrad.Gradient
	Font     *figletlib.Font
	Text     string
}

var baseConfig = LogoConfiguration{
	Style: &lipgloss.Style{},
	Text:  cases.Title(language.English, cases.Compact).String(ui.AppName),
	Font:  randomFiglet(),
}

var figlets, _ = assets.LoadAllFiglets()

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

func randomFiglet() *figletlib.Font {
	figletNames := maps.Keys(figlets)
	randomFiglet := figlets[figletNames[utils.RandomInt(len(figletNames))]]

	font, err := figletlib.ReadFontFromBytes(randomFiglet)
	if err != nil {
		panic(err)
	}

	return font
}

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

// Static returns a string representation of the logo using the provided options.
// It accepts zero or more LogoOption arguments to customize the logo.
func Static(options ...LogoOption) string {
	config := baseConfig
	return buildLogo(config, options...)
}

// PrintAnimated prints the animated logo by splitting the colorized logo into lines
// and printing each line with a slight delay between them.
func PrintAnimated() {
	config := baseConfig
	colorized := buildLogo(
		config,
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
func buildLogo(config LogoConfiguration, options ...LogoOption) string {
	for _, opt := range options {
		opt(&config)
	}

	img := figletlib.SprintMsg(config.Text, config.Font, styles.TerminalWidth(), config.Font.Settings(), "left")

	if utils.IsZero(config.Gradient) {
		return config.Style.Render(img)
	}

	return config.Style.Render(colorize(img, config.Gradient))
}

func colorize(img string, grad colorgrad.Gradient) string {
	splited := strings.Split(img, "\n")
	lines := len(splited)
	grad = grad.Sharp(uint(lines), 0)
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

	return strings.Join(colorized, "\n")
}
