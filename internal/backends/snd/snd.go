package snd

import "github.com/maxb-odessa/slog"

type Snd struct {
	name string
}

func Run(scope string) (interface{}, error) {
	slog.Debug(9, "snd Run %s", scope)
	var s Snd
	s.name = scope
	return s, nil
}

func (s Snd) Name() string {
	return s.name
}

func (s Snd) Consumer() chan interface{} {
	return make(chan interface{})
}

func (s Snd) Producer() chan interface{} {
	return make(chan interface{})
}
