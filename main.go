package main

import (
	"fmt"
	"os"
	"rapper/cli"
	"rapper/cli/ui"
	"rapper/files"
	"rapper/versions"
	"rapper/web"
)

var AppVersion = "2.3.0"
var AppName = "rapper"

func main() {
	path := handleArgs()

	config, err := files.Config(path)
	if err != nil {
		handleExit(err)
	}

	hg := web.NewHttpGateway(config.Token, config.Path.Method, config.Path.Template, config.Payload.Template)

	c, err := cli.New(config, path, hg, AppName, AppVersion)
	if err != nil {
		handleExit(err)
	}

	if err := c.Start(); err != nil {
		handleExit(err)
	}
	handleExit(nil)
}

func handleExit(err error) {
	update := versions.CheckForUpdate(web.NewHttpClient(), AppVersion)
	if err == nil {
		cli.Exit(update)
	}
	if update != versions.NoUpdates {
		update = "\n\n" + update
	}
	cli.Exit(err.Error() + update)
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
