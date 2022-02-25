package file

import (
	"gamenode/internal/pubsub"

	"github.com/maxb-odessa/slog"
)

type File struct {
	name string
}

func Init(confScope string) (interface{}, error) {
	slog.Debug(9, "file INIT %s", confScope)
	return File{name: confScope}, nil
}

func (j File) Run(broker *pubsub.Pubsub) error {
	return nil
}
