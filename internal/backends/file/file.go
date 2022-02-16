package file

import "github.com/maxb-odessa/slog"

type File struct {
}

func Run(scope string) (interface{}, error) {
	slog.Debug(9, "file Run %s", scope)
	var f File
	return f, nil
}

func (f File) Name() string {
	return "file name"
}

func (f File) Consumer() chan interface{} {
	return make(chan interface{})
}

func (f File) Producer() chan interface{} {
	return make(chan interface{})
}
