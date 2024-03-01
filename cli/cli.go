package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
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
	csv        csvOption
	ctx        context.Context
	cancelFn   context.CancelFunc
	outputFile string

	showProgress bool
	progressBar  progress.Model
	completed    float64
	errs         int

	viewport viewport.Model
	logs     []string
	logsCh   chan string

	filesList list.Model
	help      help.Model
	keyMap    keyMap

	gateway web.HttpGateway
}

func (c *Cli) Start() error {
	defer close(c.logsCh)
	if _, err := tea.NewProgram(c).Run(); err != nil {
		return err
	}
	return nil
}

func (c *Cli) logManager() {
	for log := range c.logsCh {
		c.logs = append(c.logs, log)
		c.viewport.SetContent(strings.Join(c.logs, "\n"))
		c.viewport.GotoBottom()
	}
}

func New(config files.AppConfig, path string, gateway web.HttpGateway, appName string, appVersion string, outputFile string) (*Cli, error) {
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
		outputFile:  outputFile,
		logsCh:      make(chan string),
		filesList:   createList(opts),
		progressBar: progress.New(progress.WithDefaultGradient()),
		help:        createHelp(),
		viewport:    viewport.New(20, 60),
		keyMap:      keys,
		gateway:     gateway,
	}, nil
}

func (c *Cli) Init() tea.Cmd {
	go c.logManager()
	return tea.Batch(tea.EnterAltScreen, tickCmd())
}

func (c *Cli) execRequests(ctx context.Context, file Option[string], ch chan<- string) {
	defer c.done()
	if ch != nil {
		defer close(ch)
	}
	csv, err := files.MapCSV(file.Value, c.csv.sep, c.csv.fields)
	if err != nil {
		c.logsCh <- fmtError(CSV, err.Error())
		c.cancel()
		return
	}
	total := len(csv.Lines)
	current := 0

	if total == 0 {
		c.logsCh <- fmtError(CSV, "No records found in the file\n")
		c.cancel()
		return
	}
	c.logsCh <- fmt.Sprintf("%s Processing file #%d: %s", ui.IconWomanDancing, c.filesList.Index()+1, ui.Green(file.Title))

Processing:
	for i, record := range csv.Lines {
		select {
		case <-ctx.Done():
			c.logsCh <- fmtError(Cancelation, fmt.Sprintf("Processed %d of %d", current, total))
			break Processing
		default:
			current = 1 + i
			c.completed = float64(current) / float64(total)
			response, err := c.gateway.Exec(record)
			if err != nil {
				c.logsCh <- fmtError(Request, err.Error())
				c.errs++
			} else if response.Status != http.StatusOK {
				c.logsCh <- fmtError(Status, fmtStatusError(record, response.Status))
				c.errs++
			}
			if c.outputFile != "" {
				ch <- fmt.Sprintf("url=%s status=%d error=%s body=%s", response.URL, response.Status, err, response.Body)
			}
		}
	}
	c.logsCh <- formatDoneMessage(c.errs)
}

func (c *Cli) writeOutputFile(outputFile string, ch <-chan string) {
	file, err := os.OpenFile(outputFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		c.logsCh <- fmtError(OutputFile, err.Error())
	}
	defer file.Close()
	for log := range ch {
		if _, err := file.WriteString(log + "\n"); err != nil {
			c.logsCh <- fmtError(OutputFile, err.Error())
		}
	}
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
	var ch chan string
	if c.ctx != nil {
		c.logsCh <- fmt.Sprintf("\n%s  %s\n", ui.IconInformation, "Please wait until the current operation is finished")
	} else {
		c.resetProgress()
		if c.outputFile != "" {
			ch = make(chan string)
			go c.writeOutputFile(c.outputFile, ch)
		}
		c.ctx, c.cancelFn = context.WithCancel(context.Background())
		go c.execRequests(c.ctx, item, ch)
	}
}

func (c *Cli) resizeElements(width int, height int) {
	c.filesList.SetWidth(int(float64(width) * 0.3))

	logViewWidth := width - lipgloss.Width(c.filesList.View()) - 7
	c.progressBar.Width = logViewWidth
	headerHeight := lipgloss.Height(viewPortTitle) + 9

	c.viewport = viewport.New(logViewWidth, (height - headerHeight))
	c.viewport.YPosition = headerHeight
	c.viewport.SetContent(strings.Join(c.logs, "\n"))
	c.viewport.GotoBottom()
}

func (c *Cli) logView() string {
	return fmt.Sprintf("%s\n%s\n\n", viewPortTitle, c.viewport.View())
}
