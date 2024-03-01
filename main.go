package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anibaldeboni/rapper/cli"
	"github.com/anibaldeboni/rapper/cli/ui"
	"github.com/anibaldeboni/rapper/files"
	"github.com/anibaldeboni/rapper/versions"
	"github.com/anibaldeboni/rapper/web"
	tea "github.com/charmbracelet/bubbletea"
)

var AppVersion = "2.5.2"
var AppName = "rapper"

func main() {
	cwd, _ := os.Getwd()
	configPath := flag.String("config", cwd, "path to config file")
	workingDir := flag.String("dir", cwd, "path to directory containing the CSV files")
	outputFile := flag.String("output", "", "path to output file, including the file name")
	flag.Usage = usage
	flag.Parse()

	config, err := files.Config(*configPath)
	if err != nil {
		handleExit(err)
	}

	hg := web.NewHttpGateway(config.Token, config.Path.Method, config.Path.Template, config.Payload.Template)

	c, err := cli.New(config, *workingDir, hg, AppName, AppVersion, *outputFile)
	if err != nil {
		handleExit(err)
	}
	if _, err := tea.NewProgram(c).Run(); err != nil {
		handleExit(err)
	}

	handleExit(nil)
}

func handleExit(err error) {
	update := updateCheck()
	if err == nil {
		cli.Exit(update)
	}
	if update != versions.NoUpdates {
		update = "\n\n" + update
	}
	cli.Exit(err.Error() + update)
}

func updateCheck() string {
	return versions.CheckForUpdate(web.NewHttpClient(), AppVersion)
}

func usage() {
	fmt.Printf("%s (%s)\n", ui.Bold(AppName), AppVersion)
	fmt.Println("\nA CLI tool to send HTTP requests based on CSV files.")
	fmt.Printf("All flags are optional. If %s or %s are not provided, the current directory will be used.\n", ui.Bold("-config"), ui.Bold("-dir"))
	fmt.Printf("If %s file is not provided, the responses bodies will not be saved.\n", ui.Bold("-output"))
	fmt.Println("\nUsage:")
	fmt.Printf("  %s [options]\n", ui.Bold(filepath.Base(os.Args[0])))
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\n" + updateCheck())
}
