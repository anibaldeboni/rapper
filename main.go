package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anibaldeboni/rapper/internal/config"
	"github.com/anibaldeboni/rapper/internal/logs"
	"github.com/anibaldeboni/rapper/internal/processor"
	"github.com/anibaldeboni/rapper/internal/styles"
	"github.com/anibaldeboni/rapper/internal/ui"
	"github.com/anibaldeboni/rapper/internal/ui/logo"
	"github.com/anibaldeboni/rapper/internal/updates"
	"github.com/anibaldeboni/rapper/internal/utils"
	"github.com/anibaldeboni/rapper/internal/web"
)

var (
	reqsPath   *string
	filesPath  *string
	outputFile *string
	updateMsg  = make(chan string)
)

func init() {
	go updateCheck(updateMsg)
	cwd, _ := os.Getwd()
	reqsPath = flag.String("requests", cwd, "path to directory containing a requests file")
	filesPath = flag.String("files", cwd, "path to directory containing the CSV files")
	outputFile = flag.String("output", "", "path to output file, including the file name")
	flag.Usage = usage
	flag.Parse()
}

func main() {
	runMgr, err := config.NewRunManager(*reqsPath)
	if err != nil {
		handleExit(fmt.Errorf("could not read config file: %w", err))
	}

	cfg := runMgr.Get()
	if cfg == nil {
		handleExit(errors.New("no active configuration found"))
		return
	}

	logger := logs.NewLogger(*outputFile)

	hg, err := web.NewHttpGateway(
		cfg.Request.Method,
		cfg.Request.URLTemplate,
		cfg.Request.BodyTemplate,
		cfg.Request.Headers,
	)
	if err != nil {
		handleExit(fmt.Errorf("could not create HTTP gateway: %w", err))
	}

	csvProcessor := processor.NewProcessor(
		cfg.CSV,
		hg,
		logger,
	)

	runMgr.OnChange(func(newCfg *config.Config) {
		_ = hg.UpdateConfig(
			newCfg.Request.Method,
			newCfg.Request.URLTemplate,
			newCfg.Request.BodyTemplate,
			newCfg.Request.Headers,
		)
		csvProcessor.UpdateConfig(newCfg.CSV)
	})

	filePaths, err := utils.FindFiles(*filesPath, "*.csv")
	if err != nil {
		handleExit(fmt.Errorf("could not execute file scan in %s: %w", styles.Bold(*filesPath), err))
	}
	if len(filePaths) == 0 {
		handleExit(fmt.Errorf("no CSV files found in %s", styles.Bold(*filesPath)))
	}

	tui := ui.NewApp(filePaths, csvProcessor, logger, runMgr)
	ttyIn, ttyOut, err := tea.OpenTTY()
	if err != nil {
		handleExit(fmt.Errorf("could not open terminal tty: %w", err))
	}
	defer func() {
		_ = ttyIn.Close()
		_ = ttyOut.Close()
	}()

	if _, err := tea.NewProgram(
		tui,
		tea.WithInput(ttyIn),
		tea.WithOutput(ttyOut),
	).Run(); err != nil {
		handleExit(fmt.Errorf("could not run the program: %w", err))
	}

	handleExit()
}

func usage() {
	fmt.Printf("%s (%s)\n", styles.Bold(ui.AppName), ui.AppVersion)
	fmt.Println("\nA CLI tool to send HTTP requests based on CSV files.")
	fmt.Printf("All flags are optional. If %s or %s are not provided, the current directory will be used.\n", styles.Bold("-requests"), styles.Bold("-files"))
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
