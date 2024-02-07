package cli

import (
	"context"
	"fmt"
	"github.com/anibaldeboni/rapper/files"
	"github.com/anibaldeboni/rapper/web"
	"net/http"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	name    string
	version string
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
	alert        string
	errs         []string

	filesList list.Model
	file      string
	help      help.Model
	keys      keyMap

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
		keys:        keys,
		gateway:     gateway,
	}, nil
}

func (c *Cli) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, tickCmd())
}

func (c *Cli) execRequests(ctx context.Context, filePath string) {
	csv, err := files.MapCSV(filePath, c.csv.sep, c.csv.fields)
	if err != nil {
		c.addError(formatError("CSV error", err.Error()))
		c.cancel()
		return
	}
	total := len(csv.Lines)
	current := 0

	if total == 0 {
		c.addError(formatError("CSV error", "No records found in the file"))
		c.cancel()
		return
	}

Processing:
	for i, record := range csv.Lines {
		select {
		case <-ctx.Done():
			completed := fmt.Sprintf("Processed %d of %d", current, total)
			c.addError(formatError("Operation canceled", completed))
			break Processing
		default:
			current = 1 + i
			c.completed = float64(current) / float64(total)
			response, err := c.gateway.Exec(record)
			if err != nil {
				c.addError(formatError("Request error", err.Error()))
			} else if response.Status != http.StatusOK {
				c.addError(formatStatusError(record, response.Status))
			}
		}
	}
	c.done()
}

func (c *Cli) addError(err string) {
	c.errs = append(c.errs, err)
}

func (c *Cli) cancel() {
	c.cancelFn()
	c.done()
}

func (c *Cli) done() {
	c.ctx = nil
	c.cancelFn = nil
	c.alert = ""
}

func (c *Cli) resetProgress() {
	c.showProgress = true
	c.completed = 0
	c.progressBar.SetPercent(0)
	c.errs = nil
	c.alert = ""
}

func (c *Cli) selectItem(item Option[string]) {
	if c.ctx != nil {
		c.alert = "Please wait until the current operation is finished"
	} else {
		c.resetProgress()
		c.file = item.Title
		c.ctx, c.cancelFn = context.WithCancel(context.Background())
		go c.execRequests(c.ctx, item.Value)
	}
}

func (c *Cli) resizeElements(width int) {
	listWidth := int(float64(width) * 0.4)
	progressWidth := width - listWidth + 4
	c.filesList.SetWidth(listWidth)
	c.progressBar.Width = progressWidth
}

func (c *Cli) isProcessing() bool {
	return c.ctx != nil
}
