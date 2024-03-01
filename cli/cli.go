package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"

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
	logs          = &Logs{}
	csv           csvOption
	gateway       web.HttpGateway
	outputFile    string
	completed     float64
	errs          int
	ctx           context.Context
	cancelFn      context.CancelFunc
	showProgress  bool
)

type csvOption struct {
	sep    string
	fields []string
}

type Cli struct {
	progressBar progress.Model
	viewport    viewport.Model
	filesList   list.Model
	help        help.Model
	keyMap      keyMap
	logCount    int
}

func New(config files.AppConfig, path string, hg web.HttpGateway, appName string, appVersion string, of string) (Cli, error) {
	opts, err := findCsv(path)
	if err != nil {
		return Cli{}, err
	}
	name = appName
	version = appVersion
	outputFile = of
	logs = &Logs{}
	csv = csvOption{
		sep:    config.CSV.Separator,
		fields: config.CSV.Fields,
	}
	gateway = hg

	return Cli{
		filesList:   createList(opts, "Choose a file to process"),
		progressBar: progress.New(progress.WithDefaultGradient()),
		help:        createHelp(),
		viewport:    viewport.New(20, 60),
		keyMap:      keys,
	}, nil
}

func (c Cli) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, tickCmd())
}

func execRequests(ctx context.Context, file Option[string], ch chan<- string) {
	defer done()
	if ch != nil {
		defer close(ch)
	}
	csv, err := files.MapCSV(file.Value, csv.sep, csv.fields)
	if err != nil {
		logs.Add(fmtError(CSV, err.Error()))
		cancel()
		return
	}
	total := len(csv.Lines)
	current := 0

	if total == 0 {
		logs.Add(fmtError(CSV, "No records found in the file\n"))
		cancel()
		return
	}
	logs.Add(fmt.Sprintf("%s Processing file %s", ui.IconWomanDancing, ui.Green(file.Title)))

Processing:
	for i, record := range csv.Lines {
		select {
		case <-ctx.Done():
			logs.Add(fmtError(Cancelation, fmt.Sprintf("Processed %d of %d", current, total)))
			break Processing
		default:
			current = 1 + i
			completed = float64(current) / float64(total)
			response, err := gateway.Exec(record)
			if err != nil {
				logs.Add(fmtError(Request, err.Error()))
				errs++
			} else if response.Status != http.StatusOK {
				logs.Add(fmtError(Status, fmtStatusError(record, response.Status)))
				errs++
			}
			if outputFile != "" {
				ch <- fmt.Sprintf("url=%s status=%d error=%s body=%s", response.URL, response.Status, err, response.Body)
			}
		}
	}
	logs.Add(formatDoneMessage(errs))
}

func writeOutputFile(outputFile string, ch <-chan string) {
	file, err := os.OpenFile(outputFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		logs.Add(fmtError(OutputFile, err.Error()))
	}
	defer file.Close()
	for log := range ch {
		if _, err := file.WriteString(log + "\n"); err != nil {
			logs.Add(fmtError(OutputFile, err.Error()))
		}
	}
}

func cancel() {
	cancelFn()
	done()
}

func done() {
	ctx = nil
	cancelFn = nil
	errs = 0
}

func (c Cli) selectItem(item Option[string]) Cli {
	var ch chan string
	if ctx != nil {
		logs.Add(fmt.Sprintf("\n%s  %s\n", ui.IconInformation, "Please wait until the current operation is finished"))
	} else {
		showProgress = true
		completed = 0
		c.progressBar.SetPercent(0)
		if outputFile != "" {
			ch = make(chan string)
			go writeOutputFile(outputFile, ch)
		}
		ctx, cancelFn = context.WithCancel(context.Background())
		go execRequests(ctx, item, ch)
	}
	return c
}

func (c Cli) resizeElements(width int, height int) Cli {
	logViewWidth := width - lipgloss.Width(c.filesList.View()) - 7
	headerHeight := lipgloss.Height(viewPortTitle) + 9

	c.progressBar.Width = logViewWidth
	c.viewport.Height = height - headerHeight
	c.viewport.Width = logViewWidth
	c.viewport.YPosition = headerHeight

	return c
}
