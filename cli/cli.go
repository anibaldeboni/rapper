package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anibaldeboni/rapper/cli/messages"
	"github.com/anibaldeboni/rapper/internal/log"
	"github.com/anibaldeboni/rapper/internal/processor"
	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/versions"
	"github.com/anibaldeboni/rapper/internal/web"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	AppName       = "rapper"
	AppVersion    = "2.5.2"
	viewPortTitle = styles.TitleStyle.Render("Execution logs")
	logs          = log.NewLogManager()
	ctx           context.Context
	cancel        context.CancelFunc
	state         = &State{}
	csvProcessor  processor.Processor
)

type Cli interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

type cliModel struct {
	viewport  viewport.Model
	filesList list.Model
	help      help.Model
	spinner   spinner.Model
	width     int
}

func New(csvFiles []string, csvProc processor.Processor, logManager log.LogManager) Cli {
	state.Set(SelectFile)
	csvProcessor = csvProc
	logs = logManager

	return cliModel{
		viewport:  viewport.New(20, 60),
		filesList: createList(mapListOptions(csvFiles), "Choose a file"),
		help:      createHelp(),
		spinner:   createSpinner(),
	}
}

func createSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205"))

	return s
}

func Usage() {
	fmt.Printf("%s (%s)\n", styles.Bold(AppName), AppVersion)
	fmt.Println("\nA CLI tool to send HTTP requests based on CSV files.")
	fmt.Printf("All flags are optional. If %s or %s are not provided, the current directory will be used.\n", styles.Bold("-config"), styles.Bold("-dir"))
	fmt.Printf("If %s file is not provided, the responses bodies will not be saved.\n", styles.Bold("-output"))
	fmt.Println("\nUsage:")
	fmt.Printf("  %s [options]\n", styles.Bold(filepath.Base(os.Args[0])))
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\n" + UpdateCheck())
}

func UpdateCheck() string {
	return versions.CheckForUpdate(web.NewHttpClient(), AppVersion)
}

func stop() {
	cancel()
	state.Set(Stale)
}

func setContext() {
	ctx, cancel = context.WithCancel(context.Background())
	state.Set(Running)
}

func (this cliModel) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, tickCmd(), this.spinner.Tick)
}

func (this cliModel) selectItem(item Option[string]) cliModel {
	if state.Get() != Running {
		setContext()
		csvProcessor.Do(ctx, stop, item.Value)
	} else {
		logs.Add(messages.NewOperationError())
	}
	return this
}

func (this cliModel) resizeElements(width int, height int) cliModel {
	this.width = width
	this.filesList.SetHeight(height - 4)

	logViewWidth := width - lipgloss.Width(this.filesList.View())
	headerHeight := lipgloss.Height(viewPortTitle)

	this.viewport.Height = height - headerHeight - 6
	this.viewport.Width = logViewWidth - 2

	return this
}
