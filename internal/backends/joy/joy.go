package joy

import "github.com/maxb-odessa/slog"

type Joy struct{}

func Run(scope string) (interface{}, error) {
	slog.Debug(9, "joy Run")
	var j Joy
	return j, nil
}

func (j Joy) Name() string {
	return "file name"
}

func (j Joy) Consumer() chan interface{} {
	return make(chan interface{})
}

func (j Joy) Producer() chan interface{} {
	return make(chan interface{})
}
