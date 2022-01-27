package main

import (
	"github.com/pborman/getopt/v2"

	"ednode/pkg/config"
	"ednode/pkg/slog"
)

func main() {

	// get cmdline args and parse them
	help := false
	debug := 0
	configFile := "./ednode.conf"
	getopt.HelpColumn = 0
	getopt.FlagLong(&help, "help", 'h', "Show this help")
	getopt.FlagLong(&debug, "debug", 'd', "Set debug level")
	getopt.FlagLong(&configFile, "config", 'c', "Path to config file")
	getopt.Parse()

	// help-only requested
	if help {
		getopt.Usage()
		return
	}

	// setup logger
	slog.Init("ednode", debug, "2006-01-02 15:04:05")

	// read config file
	if err := config.Read(configFile); err != nil {
		slog.Fatal("config failed: %s", err)
	}

	// start sources readers
}
