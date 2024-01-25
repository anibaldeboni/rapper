package main

import (
	"fmt"
	"os"
	"rapper/cli"
	"rapper/files"
	"rapper/ui"
	"rapper/web"
)

var AppVersion = "0.0.1"
var AppName = "rapper"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "help" {
		usage()
		os.Exit(0)
	}
	path, err := os.Getwd()
	if err != nil {
		cli.ExitOnError(err.Error())
	}

	config, err := files.Config(path)
	if err != nil {
		cli.ExitOnError(err.Error())
	}

	csvPath, err := files.ChooseFile(path)
	if err != nil {
		cli.ExitOnError(err.Error())
	}

	csv, err := files.MapCSV(csvPath)
	if err != nil {
		cli.ExitOnError(err.Error())
	}

	filteredCSV := files.FilterCSV(csv, config.CSV)
	hg := web.NewHttpGateway(config.Token, config.Path.Method, config.Path.Template, config.Payload.Template)

	if err := cli.Run(filteredCSV, hg); err != nil {
		cli.ExitOnError(err.Error())
	}
}

func usage() {
	fmt.Printf("%s (%s)\nUsage: %s (in a folder containing a config.yml file)\n", ui.Bold(AppName), AppVersion, AppName)
}
