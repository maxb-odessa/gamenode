package main

import (
	"github.com/maxb-odessa/gamenode/internal/router"

	"github.com/pborman/getopt/v2"

	"github.com/maxb-odessa/sconf"
	"github.com/maxb-odessa/slog"
)

func main() {

	// get cmdline args and parse them
	help := false
	debug := 0
	configFile := "etc/gamenode.conf"
	getopt.HelpColumn = 0
	getopt.FlagLong(&help, "help", 'h', "Show this help")
	getopt.FlagLong(&debug, "debug", 'd', "Set debug log level")
	getopt.FlagLong(&configFile, "config", 'c', "Path to config file")
	getopt.Parse()

	// help-only requested
	if help {
		getopt.Usage()
		return
	}

	// setup logger
	slog.Init("gamenode", debug, "2006-01-02 15:04:05")
	slog.Info("started")

	// read config file
	if err := sconf.Read(configFile); err != nil {
		slog.Fatal("config failed: %s", err)
	}

	// run router
	if err := router.Start(); err != nil {
		slog.Fatal("failed to start router: %s", err)
	}

	return
}
