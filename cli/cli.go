package cli

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/anibaldeboni/rapper/cli/log"
	"github.com/anibaldeboni/rapper/cli/messages"
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
	viewPortTitle = ui.TitleStyle.Render("Execution logs")
	logs          = &log.Logs{}
	csvSeparator  string
	csvFields     []string
	gateway       web.HttpGateway
	outputStream  output.Output
	completed     float64
	ctx           context.Context
	cancel        context.CancelFunc
	state         = &State{}
)

type Cli interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

type cliModel struct {
	progressBar progress.Model
	viewport    viewport.Model
	filesList   list.Model
	help        help.Model
}

func New(config files.AppConfig, path string, hg web.HttpGateway, of string) (Cli, error) {
	opts, err := findCsv(path)
	if err != nil {
		return cliModel{}, err
	}

	state.Set(SelectFile)
	outputStream = output.New(of, logs)
	csvSeparator = config.CSV.Separator
	csvFields = config.CSV.Fields
	gateway = hg

	go outputStream.Listen()

	return cliModel{
		filesList:   createList(opts, "Choose a file"),
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

func execRequests(ctx context.Context, file Option[string], gateway web.HttpGateway, outputStream output.Output, logs *log.Logs) {
	defer stop()
	csv, err := files.MapCSV(file.Value, csvSeparator, csvFields)
	if err != nil {
		logs.Add(messages.NewCsvError(err.Error()))
		return
	}

	if len(csv.Lines) == 0 {
		logs.Add(messages.NewCsvError("No records found in the file\n"))
		return
	}

	logs.Add(messages.NewProcessingMessage(file.Title))
	errs := 0
	total := len(csv.Lines)
	current := 0

Processing:
	for i, record := range csv.Lines {
		select {
		case <-ctx.Done():
			logs.Add(messages.NewCancelationError(fmt.Sprintf("Processed %d of %d", current, total)))
			break Processing
		default:
			current = 1 + i
			completed = float64(current) / float64(total)
			response, err := gateway.Exec(record)
			if err != nil {
				logs.Add(messages.NewRequestError(err.Error()))
				errs++
			} else if response.Status != http.StatusOK {
				logs.Add(messages.NewHttpStatusError(record, response.Status))
				errs++
			}
			outputStream.Send(output.NewMessage(response.URL, response.Status, err, response.Body))
		}
	}
	logs.Add(messages.NewDoneMessage(errs))
}

func stop() {
	cancel()
	state.Set(Stale)
}

func setContext() {
	ctx, cancel = context.WithCancel(context.Background())
	state.Set(Running)
}

func (c cliModel) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, tickCmd())
}

func (c cliModel) selectItem(item Option[string]) cliModel {
	if state.Get() != Running {
		setContext()
		completed = 0
		c.progressBar.SetPercent(0)
		go execRequests(ctx, item, gateway, outputStream, logs)
	} else {
		logs.Add(messages.NewOperationError())
	}
	return c
}

func (c cliModel) resizeElements(width int, height int) cliModel {
	c.filesList.SetHeight(height - 4)

	logViewWidth := width - lipgloss.Width(c.filesList.View())
	headerHeight := lipgloss.Height(viewPortTitle)

	c.progressBar.Width = logViewWidth - 6

	c.viewport.Height = height - headerHeight - 8
	c.viewport.Width = logViewWidth - 2

	return c
}
