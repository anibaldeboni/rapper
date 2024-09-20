package logo

import (
	"fmt"
	"strings"
	"time"

	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/ui/assets"
	"github.com/anibaldeboni/rapper/internal/utils"
	"github.com/ccoveille/go-safecast"
	"github.com/charmbracelet/lipgloss"
	"github.com/mazznoer/colorgrad"
	"github.com/michaelquigley/figlet/figletlib"
	"golang.org/x/exp/maps"
)

type Option func(*Config)

type Config struct {
	Style           *lipgloss.Style
	Gradient        colorgrad.Gradient
	ColoringPattern func(string, colorgrad.Gradient) string
	Font            *figletlib.Font
	Text            string
}

var (
	figlets    assets.Figlets
	baseConfig Config
)

func rndLogo() string {
	weightedNames := map[string]float64{
		"Rapper": 0.9,
		"Aggro!": 0.1,
	}

	return utils.WeightedRandom(weightedNames)
}

func init() {
	figlets, _ = assets.LoadAllFiglets()
	baseConfig = Config{
		Style:           &lipgloss.Style{},
		Text:            rndLogo(),
		Font:            randomFiglet(),
		ColoringPattern: horizontalColoring,
	}
}

func randomFiglet() *figletlib.Font {
	figletNames := maps.Keys(figlets)
	randomFiglet := figlets[figletNames[utils.RandomInt(len(figletNames))]]

	font, err := figletlib.ReadFontFromBytes(randomFiglet)
	if err != nil {
		panic(fmt.Errorf("Error reading font: %w", err))
	}

	return font
}

// Static returns a string representation of the logo using the provided options.
// It accepts zero or more Option arguments to customize the logo.
func Static(options ...Option) string {
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
		WithDiagonalColoring(),
	)

	lines := strings.Split(colorized, "\n")

	fmt.Print("\n\n")

	for _, line := range lines {
		fmt.Println(line)
		time.Sleep(50 * time.Millisecond)
	}
}

func buildLogo(config Config, options ...Option) string {
	for _, opt := range options {
		opt(&config)
	}

	img := figletlib.SprintMsg(config.Text, config.Font, styles.TerminalWidth(), config.Font.Settings(), "left")

	if utils.IsZero(config.Gradient) {
		return config.Style.Render(img)
	}

	return config.Style.Render(config.ColoringPattern(img, config.Gradient))
}

func horizontalColoring(str string, grad colorgrad.Gradient) string {
	splited := strings.Split(str, "\n")
	lines := len(splited)
	grad = grad.Sharp(uint(lines), 0)
	steps := 1.0 / float64(lines)

	var colorized []string
	for i, line := range splited {
		color := grad.At(steps * float64(i))
		colorized = append(
			colorized,
			lipgloss.NewStyle().
				Foreground(lipgloss.Color(color.HexString())).
				Render(line),
		)
	}

	return strings.Join(colorized, "\n")
}

func diagonalColoring(str string, grad colorgrad.Gradient) string {
	lines := strings.Split(str, "\n")
	numVerticals, err := safecast.ToUint(len(lines) + len(lines[0]) - 1)
	if err != nil {
		numVerticals = 2
	}
	grad = grad.Sharp(numVerticals, 0)
	step := 1.0 / float64(numVerticals)

	var colorized []string
	for i, line := range lines {
		chars := strings.Split(line, "")
		for j, char := range chars {
			color := grad.At(step * float64(i+j))
			chars[j] = lipgloss.NewStyle().
				Foreground(lipgloss.Color(color.HexString())).
				Render(char)
		}
		colorized = append(colorized, strings.Join(chars, ""))
	}

	return strings.Join(colorized, "\n")
}
