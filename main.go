package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/processor"
	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/ui"
	"github.com/anibaldeboni/rapper/internal/ui/logo"
	"github.com/anibaldeboni/rapper/internal/updates"
	"github.com/anibaldeboni/rapper/internal/utils"
	"github.com/anibaldeboni/rapper/internal/web"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	configPath *string
	workingDir *string
	outputFile *string
	workers    *int
	updateMsg  = make(chan string)
)

func init() {
	go updateCheck(updateMsg)
	cwd, _ := os.Getwd()
	configPath = flag.String("config", cwd, "path to directory containing a config file")
	workingDir = flag.String("dir", cwd, "path to directory containing the CSV files")
	outputFile = flag.String("output", "", "path to output file, including the file name")
	workers = flag.Int("workers", 1, fmt.Sprintf("number of request workers (max: %d)", processor.MaxWorkers))
	flag.Usage = usage
	flag.Parse()
}

func main() {
	// Create config manager (supports multi-profile)
	configMgr, err := config.NewManager(*configPath)
	if err != nil {
		handleExit(fmt.Errorf("could not read config file: %w", err))
	}

	// Get active configuration
	cfg := configMgr.Get()
	if cfg == nil {
		handleExit(errors.New("no active configuration found"))
		return // handleExit calls os.Exit, but this helps the linter
	}

	logger := logs.NewLoggger(*outputFile)

	// Create HTTP gateway with flexible headers
	hg, err := web.NewHttpGateway(
		cfg.Request.Method,
		cfg.Request.URLTemplate,
		cfg.Request.BodyTemplate,
		cfg.Request.Headers,
	)
	if err != nil {
		handleExit(fmt.Errorf("could not create HTTP gateway: %w", err))
	}

	// Use workers from config if available, otherwise use flag
	workerCount := *workers
	if cfg.Workers > 0 {
		workerCount = utils.Clamp(cfg.Workers, 1, processor.MaxWorkers)
	}

	csvProcessor := processor.NewProcessor(
		cfg.CSV,
		hg,
		logger,
		workerCount,
	)

	// Register config change listener to update gateway
	configMgr.OnChange(func(newCfg *config.Config) {
		_ = hg.UpdateConfig(
			newCfg.Request.Method,
			newCfg.Request.URLTemplate,
			newCfg.Request.BodyTemplate,
			newCfg.Request.Headers,
		)
	})

	filePaths, err := utils.FindFiles(*workingDir, "*.csv")
	if err != nil {
		handleExit(fmt.Errorf("could not execute file scan in %s: %w", styles.Bold(*workingDir), err))
	}
	if len(filePaths) == 0 {
		handleExit(fmt.Errorf("no CSV files found in %s", styles.Bold(*workingDir)))
	}

	// Use new AppModel with multi-view support
	tui := ui.NewApp(filePaths, csvProcessor, logger, configMgr)

	if _, err := tea.NewProgram(tui).Run(); err != nil {
		handleExit(fmt.Errorf("could not run the program: %w", err))
	}

	handleExit()
}

func usage() {
	fmt.Printf("%s (%s)\n", styles.Bold(ui.AppName), ui.AppVersion)
	fmt.Println("\nA CLI tool to send HTTP requests based on CSV files.")
	fmt.Printf("All flags are optional. If %s or %s are not provided, the current directory will be used.\n", styles.Bold("-config"), styles.Bold("-dir"))
	fmt.Printf("If %s file is not provided, the request responses will not be saved.\n", styles.Bold("-output"))
	fmt.Println("\nUsage:")
	fmt.Printf("  %s [options]\n", styles.Bold(filepath.Base(os.Args[0])))
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\n", <-updateMsg)
}

func updateCheck(updateMsg chan<- string) {
	if details, hasUpdate := updates.CheckFor(ui.AppVersion); hasUpdate {
		updateMsg <- styles.IconInformation + " New version available: " + styles.Bold(details.Version) + " (" + details.Url + ")"
	} else {
		updateMsg <- "is up-to-date."
	}
}

func handleExit(err ...error) {
	logo.PrintAnimated()

	var exitMsg = []string{<-updateMsg}
	var exitCode int

	if err != nil {
		exitCode = 1
		exitMsg = append(exitMsg, errors.Join(err...).Error())
	}

	fmt.Println(
		styles.ScreenCenteredStyle(
			strings.Join(exitMsg, "\n"),
		),
	)
	os.Exit(exitCode)
}
