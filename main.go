package main

import (
	"fmt"
	"os"
	"rapper/cli"
	"rapper/files"
	"rapper/ui"
	"rapper/web"
)

var AppVersion = "local"
var AppName = "rapper"

func main() {
	path := handleArgs()

	config, err := files.Config(path)
	if err != nil {
		cli.Exit(err)
	}

	csvPath, err := files.ChooseFile(path)
	if err != nil {
		cli.Exit(err)
	}

	csv, err := files.MapCSV(csvPath, config.CSV.Separator)
	if err != nil {
		cli.Exit(err)
	}

	filteredCSV := files.FilterCSV(csv, config.CSV.Fields)
	hg := web.NewHttpGateway(config.Token, config.Path.Method, config.Path.Template, config.Payload.Template)

	if err := cli.Run(filteredCSV, hg); err != nil {
		cli.Exit(err)
	}
}

func handleArgs() string {
	cwd, err := os.Getwd()
	if err != nil {
		cli.Exit(err)
	}
	if len(os.Args) > 1 {
		arg := os.Args[1]
		switch arg {
		case "help":
			cli.Exit(usage())
		case "version":
			cli.Exit(AppVersion)
		}
		if files.IsDir(arg) {
			return arg
		} else {
			cli.Exit(fmt.Errorf("%s is not a directory", ui.Bold(arg)))
		}
	}

	return cwd
}

func usage() string {
	return fmt.Sprintf(
		"%s (%s)\n\n"+
			"You must always have a %s file in the folder you will run the app.\n"+
			"If you don't point to a specific directory the current one will be used.\n\n"+
			"Usage: %s\n",
		ui.Bold(AppName),
		AppVersion,
		ui.Italic("config.yml"),
		ui.Bold(AppName+" [<folder-with-csv>]"),
	)
}
