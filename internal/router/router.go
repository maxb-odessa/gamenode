package router

import (
	"gamenode/internal/backends"
	"gamenode/internal/network"

	"github.com/maxb-odessa/slog"
)

func Run() error {

	// start backends
	backs, err := backends.Start()
	if err != nil {
		return err
	}
	slog.Debug(9, "started backends: %+v", backs)

	return network.Start()
}
