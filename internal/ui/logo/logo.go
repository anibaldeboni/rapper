package logo

import (
	"fmt"
	"strings"
	"time"

	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/ui/assets"
	"github.com/anibaldeboni/rapper/internal/utils"
	"github.com/charmbracelet/lipgloss"
	"github.com/mazznoer/colorgrad"
	"github.com/michaelquigley/figlet/figletlib"
	"golang.org/x/exp/maps"
)

type LogoOption func(*LogoConfiguration)

type LogoConfiguration struct {
	Style    *lipgloss.Style
	Gradient colorgrad.Gradient
	Font     *figletlib.Font
	Text     string
}

var (
	figlets    assets.Figlets
	baseConfig LogoConfiguration
)

func init() {
	figlets, _ = assets.LoadAllFiglets()
	baseConfig = LogoConfiguration{
		Style: &lipgloss.Style{},
		Text:  "Rapper",
		Font:  randomFiglet(),
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
