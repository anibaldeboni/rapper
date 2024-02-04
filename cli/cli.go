package cli

import (
	"fmt"
	"net/http"
	"path/filepath"
	"rapper/cli/ui"
	"rapper/files"
	"rapper/web"
	"sort"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	teaProgram *tea.Program
	name       string
	version    string
)

type Cli struct {
	config       files.AppConfig
	progressBar  progress.Model
	filesList    list.Model
	help         help.Model
	keys         keyMap
	errs         []string
	file         string
	showProgress bool
	gateway      web.HttpGateway
}

func (c Cli) Start() error {
	teaProgram = tea.NewProgram(c)
	if _, err := teaProgram.Run(); err != nil {
		return err
	}
	return nil
}

func New(config files.AppConfig, path string, gateway web.HttpGateway, appName string, appVersion string) (Cli, error) {
	opts, err := findCsv(path)
	if err != nil {
		return Cli{}, err
	}
	name = appName
	version = appVersion

	return Cli{
		config:      config,
		gateway:     gateway,
		filesList:   createList(opts),
		progressBar: progress.New(progress.WithDefaultGradient()),
		help:        createHelp(),
		keys:        keys,
	}, nil
}

func (c Cli) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (c *Cli) execRequests(filePath string) {
	csv, err := files.MapCSV(filePath, c.config.CSV.Separator, c.config.CSV.Fields)
	if err != nil {
		teaProgram.Send(errorMsg(fmt.Sprintf("%s [%s] %s", ui.IconSkull, ui.Bold("CSV error"), err.Error())))
	}
	total := len(csv.Lines)

	if total == 0 {
		teaProgram.Send(errorMsg(fmt.Sprintf("%s [%s] %s", ui.IconSkull, ui.Bold("CSV error"), "No lines found")))
		teaProgram.Send(progressMsg(1.0))
	}

	for i, record := range csv.Lines {
		response, err := c.gateway.Exec(record)
		if err != nil {
			teaProgram.Send(errorMsg(fmt.Sprintf("%s [%s] %s", ui.IconSkull, ui.Bold("Connection error"), err.Error())))
		} else if response.Status != http.StatusOK {
			teaProgram.Send(errorMsg(formatErrorMsg(record, response.Status)))
		}
		teaProgram.Send(progressMsg(float64(i+1) / float64(total)))
	}
}

func formatErrorMsg(record map[string]string, status int) string {
	result := ui.IconWarning + "  "
	keys := make([]string, 0, len(record))
	for k := range record {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		result += fmt.Sprintf("%s: %s ", ui.Bold(key), record[key])
	}
	result += fmt.Sprintf("status: %s", ui.Pink(fmt.Sprint(status)))

	return result
}

func findCsv(path string) ([]Option[string], error) {
	filePaths, err := files.FindFiles(path, "*.csv")
	if len(err) > 0 {
		return nil, fmt.Errorf("Could not execute file scan in %s", ui.Bold(path))
	}
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("No CSV files found in %s", ui.Bold(path))
	}

	opts := make([]Option[string], 0)
	for _, filePath := range filePaths {
		opts = append(
			opts,
			Option[string]{
				Title: filepath.Base(filePath),
				Value: filePath,
			},
		)
	}

	return opts, nil
}
