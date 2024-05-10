package main

import (
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
	"github.com/anibaldeboni/rapper/internal/utils"
	"github.com/anibaldeboni/rapper/internal/versions"
	"github.com/anibaldeboni/rapper/internal/web"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	configPath    *string
	workingDir    *string
	outputFile    *string
	workers       *int
	updateChecker versions.UpdateChecker
)

func init() {
	cwd, _ := os.Getwd()
	configPath = flag.String("config", cwd, "path to directory containing a config file")
	workingDir = flag.String("dir", cwd, "path to directory containing the CSV files")
	outputFile = flag.String("output", "", "path to output file, including the file name")
	workers = flag.Int("workers", 1, fmt.Sprintf("number of request workers (max: %d)", processor.MAX_WORKERS))
	flag.Usage = usage
	flag.Parse()
	updateChecker = versions.NewUpdateChecker(web.NewHttpClient(), ui.AppVersion)
}

func main() {
	cfg, err := config.Config(*configPath)
	if err != nil {
		handleExit(err)
	}

	logger := logs.NewLoggger(*outputFile)

	hg := web.NewHttpGateway(
		cfg.Token,
		cfg.Path.Method,
		cfg.Path.Template,
		cfg.Payload.Template,
	)

	csvProcessor := processor.NewProcessor(
		cfg.CSV,
		hg,
		logger,
		*workers,
	)

	filePaths, errs := utils.FindFiles(*workingDir, "*.csv")
	if len(errs) > 0 {
		handleExit(fmt.Errorf("Could not execute file scan in %s", styles.Bold(*workingDir)))
	}
	if len(filePaths) == 0 {
		handleExit(fmt.Errorf("No CSV files found in %s", styles.Bold(*workingDir)))
	}

	tui := ui.New(filePaths, csvProcessor, logger)

	if _, err := tea.NewProgram(tui).Run(); err != nil {
		handleExit(err)
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
	fmt.Println("\n" + updateCheck())
}

func updateCheck() string {
	update := updateChecker.CheckForUpdate()
	if update.HasUpdate() {
		return styles.IconInformation + " New version available: " + styles.Bold(update.Version()) + " (" + update.Url() + ")"
	}
	return "is up-to-date."
}

func handleExit(err ...error) {
	var exitMsg []string
	var exitCode int

	wait := make(chan bool, 1)

	go func(wait chan bool) {
		exitMsg = append(exitMsg, updateCheck())
		wait <- true
	}(wait)

	logo.PrintAnimated()
	<-wait

	if err != nil {
		exitCode = 1
		for _, e := range err {
			exitMsg = append(exitMsg, e.Error())
		}
	}

	fmt.Println(
		styles.ScreenCenteredStyle(
			strings.Join(exitMsg, "\n"),
		),
	)
	os.Exit(exitCode)
}
