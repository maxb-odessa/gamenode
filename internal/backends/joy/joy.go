package joy

import "github.com/maxb-odessa/slog"

type Joy struct {
	name string
}

func Run(scope string) (interface{}, error) {
	slog.Debug(9, "joy Run %s", scope)
	var j Joy
	j.name = scope
	return j, nil
}

func (j Joy) Name() string {
	return j.name
}

func (j Joy) Consumer() chan interface{} {
	return make(chan interface{})
}

func (j Joy) Producer() chan interface{} {
	return make(chan interface{})
}
