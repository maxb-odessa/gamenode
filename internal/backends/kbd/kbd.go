package kbd

import "github.com/maxb-odessa/slog"

type Kbd struct {
	name string
}

func Run(scope string) (interface{}, error) {
	slog.Debug(9, "kbd Run %s", scope)
	var k Kbd
	k.name = scope
	return k, nil
}

func (k Kbd) Name() string {
	return k.name
}

func (k Kbd) Consumer() chan interface{} {
	return make(chan interface{})
}

func (k Kbd) Producer() chan interface{} {
	return make(chan interface{})
}
