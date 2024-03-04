package main

import (
	"flag"
	"os"

	"github.com/anibaldeboni/rapper/cli"
	"github.com/anibaldeboni/rapper/files"
	"github.com/anibaldeboni/rapper/versions"
	"github.com/anibaldeboni/rapper/web"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cwd, _ := os.Getwd()
	configPath := flag.String("config", cwd, "path to directory containing a config file")
	workingDir := flag.String("dir", cwd, "path to directory containing the CSV files")
	outputFile := flag.String("output", "", "path to output file, including the file name")
	flag.Usage = cli.Usage
	flag.Parse()

	config, err := files.Config(*configPath)
	if err != nil {
		handleExit(err)
	}

	hg := web.NewHttpGateway(config.Token, config.Path.Method, config.Path.Template, config.Payload.Template)

	c, err := cli.New(config, *workingDir, hg, *outputFile)
	if err != nil {
		handleExit(err)
	}
	if _, err := tea.NewProgram(c).Run(); err != nil {
		handleExit(err)
	}

	handleExit(nil)
}

func handleExit(err error) {
	update := cli.UpdateCheck()
	if err == nil {
		cli.Exit(update)
	}
	if update != versions.NoUpdates {
		update = "\n\n" + update
	}
	cli.Exit(err.Error() + update)
}
