package main

import (
	"fmt"
	"os"

	"mosaic/internal/help"
	"mosaic/internal/mosaic"
)

func main() {
	if len(os.Args) == 1 || os.Args[1] == "help" {
		help.PrintHelp()
		os.Exit(0)
	}

	if os.Args[1] == "version" {
		help.PrintVersion()
		os.Exit(0)
	}

	app := new(mosaic.App)

	app.Config.Listen = ":8004"
	app.Config.Threads = 10
	app.Config.Images = 4
	app.Config.Refresh = 10

	if err := app.Config.Load(os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "config error: %s", err)
		os.Exit(1)
	}

	app.Start()
}
