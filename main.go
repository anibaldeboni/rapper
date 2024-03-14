package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/anibaldeboni/rapper/internal"
	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/execlog"
	"github.com/anibaldeboni/rapper/internal/filelogger"
	"github.com/anibaldeboni/rapper/internal/processor"
	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/ui"
	"github.com/anibaldeboni/rapper/internal/versions"
	"github.com/anibaldeboni/rapper/internal/web"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cwd, _ := os.Getwd()
	configPath := flag.String("config", cwd, "path to directory containing a config file")
	workingDir := flag.String("dir", cwd, "path to directory containing the CSV files")
	outputFile := flag.String("output", "", "path to output file, including the file name")
	workers := flag.Int("workers", 1, fmt.Sprintf("number of request workers (max: %d)", processor.MAX_WORKERS))
	flag.Usage = ui.Usage
	flag.Parse()

	cfg, err := config.Config(*configPath)
	if err != nil {
		handleExit(err)
	}

	logsManager := execlog.NewLogManager()

	hg := web.NewHttpGateway(
		cfg.Token,
		cfg.Path.Method,
		cfg.Path.Template,
		cfg.Payload.Template,
	)

	csvProcessor := processor.New(
		cfg.CSV,
		hg,
		filelogger.New(*outputFile, logsManager),
		logsManager,
		*workers,
	)

	filePaths, errs := internal.FindFiles(*workingDir, "*.csv")
	if len(errs) > 0 {
		handleExit(fmt.Errorf("Could not execute file scan in %s", styles.Bold(*workingDir)))
	}
	if len(filePaths) == 0 {
		handleExit(fmt.Errorf("No CSV files found in %s", styles.Bold(*workingDir)))
	}

	tui := ui.New(filePaths, csvProcessor, logsManager)

	if _, err := tea.NewProgram(tui).Run(); err != nil {
		handleExit(err)
	}

	handleExit(nil)
}

func handleExit(err error) {
	update := ui.UpdateCheck()
	if err == nil {
		ui.Exit(update)
	}
	if update != versions.NoUpdates {
		update = "\n\n" + update
	}
	ui.Exit(err.Error() + update)
}
