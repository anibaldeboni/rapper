package cli

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/anibaldeboni/rapper/cli/log"
	"github.com/anibaldeboni/rapper/cli/output"
	"github.com/anibaldeboni/rapper/cli/ui"
	"github.com/anibaldeboni/rapper/files"
	"github.com/anibaldeboni/rapper/versions"
	"github.com/anibaldeboni/rapper/web"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	AppName       = "rapper"
	AppVersion    = "2.5.2"
	viewPortTitle = ui.Bold("Execution logs\n")
	logs          = &log.Logs{}
	csvSeparator  string
	csvFields     []string
	gateway       web.HttpGateway
	outputStream  output.Output
	completed     float64
	errs          int
	ctx           context.Context
	cancelFn      context.CancelFunc
	showProgress  bool
)

type Cli interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

type cliImpl struct {
	progressBar progress.Model
	viewport    viewport.Model
	filesList   list.Model
	help        help.Model
}

func New(config files.AppConfig, path string, hg web.HttpGateway, of string) (Cli, error) {
	opts, err := findCsv(path)
	if err != nil {
		return cliImpl{}, err
	}

	outputStream = output.New(of, logs)
	csvSeparator = config.CSV.Separator
	csvFields = config.CSV.Fields
	gateway = hg

	return cliImpl{
		filesList:   createList(opts, "Choose a file to process"),
		progressBar: progress.New(progress.WithDefaultGradient()),
		help:        createHelp(),
		viewport:    viewport.New(20, 60),
	}, nil
}

func Usage() {
	fmt.Printf("%s (%s)\n", ui.Bold(AppName), AppVersion)
	fmt.Println("\nA CLI tool to send HTTP requests based on CSV files.")
	fmt.Printf("All flags are optional. If %s or %s are not provided, the current directory will be used.\n", ui.Bold("-config"), ui.Bold("-dir"))
	fmt.Printf("If %s file is not provided, the responses bodies will not be saved.\n", ui.Bold("-output"))
	fmt.Println("\nUsage:")
	fmt.Printf("  %s [options]\n", ui.Bold(filepath.Base(os.Args[0])))
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\n" + UpdateCheck())
}

func UpdateCheck() string {
	return versions.CheckForUpdate(web.NewHttpClient(), AppVersion)
}

func execRequests(ctx context.Context, file Option[string]) {
	defer done()
	csv, err := files.MapCSV(file.Value, csvSeparator, csvFields)
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
			outputStream.Send(output.NewMessage(response.URL, response.Status, err, response.Body))
		}
	}
	logs.Add(formatDoneMessage(errs))
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

func (c cliImpl) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, tickCmd())
}

func (c cliImpl) selectItem(item Option[string]) cliImpl {
	if ctx != nil {
		logs.Add(fmt.Sprintf("\n%s  %s\n", ui.IconInformation, "Please wait until the current operation is finished"))
	} else {
		showProgress = true
		completed = 0
		c.progressBar.SetPercent(0)
		if outputStream.Enabled() {
			go outputStream.WriteToFile()
		}
		ctx, cancelFn = context.WithCancel(context.Background())
		go execRequests(ctx, item)
	}
	return c
}

func (c cliImpl) resizeElements(width int, height int) cliImpl {
	logViewWidth := width - lipgloss.Width(c.filesList.View()) - 7
	headerHeight := lipgloss.Height(viewPortTitle) + 10

	c.progressBar.Width = logViewWidth
	c.viewport.Height = height - headerHeight
	c.viewport.Width = logViewWidth
	c.viewport.YPosition = headerHeight

	return c
}
