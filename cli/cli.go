package cli

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/anibaldeboni/rapper/cli/ui"
	"github.com/anibaldeboni/rapper/files"
	"github.com/anibaldeboni/rapper/web"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	name          string
	version       string
	viewPortTitle = ui.Bold("Execution logs\n")
)

type csvOption struct {
	sep    string
	fields []string
}

type Cli struct {
	csv      csvOption
	ctx      context.Context
	cancelFn context.CancelFunc

	showProgress bool
	progressBar  progress.Model
	completed    float64
	errs         int

	viewport viewport.Model
	logs     []string

	filesList list.Model
	file      string
	help      help.Model
	keyMap    keyMap

	gateway web.HttpGateway
}

func (c *Cli) Start() error {
	if _, err := tea.NewProgram(c).Run(); err != nil {
		return err
	}
	return nil
}

func New(config files.AppConfig, path string, gateway web.HttpGateway, appName string, appVersion string) (*Cli, error) {
	opts, err := findCsv(path)
	if err != nil {
		return &Cli{}, err
	}
	name = appName
	version = appVersion

	return &Cli{
		csv: csvOption{
			sep:    config.CSV.Separator,
			fields: config.CSV.Fields,
		},
		filesList:   createList(opts),
		progressBar: progress.New(progress.WithDefaultGradient()),
		help:        createHelp(),
		viewport:    viewport.New(20, 60),
		keyMap:      keys,
		gateway:     gateway,
	}, nil
}

func (c Cli) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, tickCmd())
}

func (c *Cli) execRequests(ctx context.Context, filePath string) {
	defer c.done()
	csv, err := files.MapCSV(filePath, c.csv.sep, c.csv.fields)
	if err != nil {
		c.addLog(fmtError(CSV, err.Error()))
		c.cancel()
		return
	}
	total := len(csv.Lines)
	current := 0

	if total == 0 {
		c.addLog(fmtError(CSV, "No records found in the file\n"))
		c.cancel()
		return
	}
	c.addLog(fmt.Sprintf("%s Processing %s", ui.IconWomanDancing, ui.Green(c.file)))

Processing:
	for i, record := range csv.Lines {
		select {
		case <-ctx.Done():
			c.addLog(fmtError(Cancelation, fmt.Sprintf("Processed %d of %d", current, total)))
			break Processing
		default:
			current = 1 + i
			c.completed = float64(current) / float64(total)
			response, err := c.gateway.Exec(record)
			if err != nil {
				c.addLog(fmtError(Request, err.Error()))
				c.errs++
			} else if response.Status != http.StatusOK {
				c.addLog(fmtError(Status, fmtStatusError(record, response.Status)))
				c.errs++
			}
		}
	}
	c.addLog(formatDoneMessage(c.errs))
}

func (c *Cli) addLog(err string) {
	c.logs = append(c.logs, err)
}

func (c *Cli) cancel() {
	c.cancelFn()
	c.done()
}

func (c *Cli) done() {
	c.ctx = nil
	c.cancelFn = nil
	c.errs = 0
}

func (c *Cli) resetProgress() {
	c.showProgress = true
	c.completed = 0
	c.progressBar.SetPercent(0)
}

func (c *Cli) selectItem(item Option[string]) {
	if c.ctx != nil {
		c.addLog(fmt.Sprintf("\n%s  %s\n", ui.IconInformation, "Please wait until the current operation is finished"))
	} else {
		c.resetProgress()
		c.file = item.Title
		c.ctx, c.cancelFn = context.WithCancel(context.Background())
		go c.execRequests(c.ctx, item.Value)
	}
}

func (c *Cli) resizeElements(width int, height int) {
	listWidth := int(float64(width) * 0.4)
	progressWidth := width - listWidth + 4
	c.filesList.SetWidth(listWidth)
	c.progressBar.Width = progressWidth

	headerHeight := lipgloss.Height(viewPortTitle)

	c.viewport = viewport.New(width, (height-headerHeight)/2)
	c.viewport.YPosition = headerHeight
	c.viewport.SetContent(strings.Join(c.logs, "\n"))
	c.viewport.GotoBottom()
}

func (c *Cli) logView() string {
	return fmt.Sprintf("%s\n%s\n\n", viewPortTitle, c.viewport.View())
}
